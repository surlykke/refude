const electron = require('electron')
const app = electron.app
console.log(app.getPath('userData'))
const secondInst = app.makeSingleInstance((argv, workingDirectory) => {
});

if (secondInst) {
    console.log("Second inst, quitting")
    app.quit()
}


const fs = require('fs')
const BrowserWindow = electron.BrowserWindow

let windows = {}

app.on('window-all-closed', function () {
})





let createWin = function(appName) {
    let settingsPath = app.getPath('userData') + "/" + appName + "-settings.json" 
   
    let saveSettings = function(settings) {
        fd = fs.openSync(settingsPath, "w") 
        fs.writeFileSync(fd, JSON.stringify(settings))
    }

    let timeoutId = null;
    let boundsChanged = function() {
        if (timeoutId) {
            clearTimeout(timeoutId)
        }
        timeoutId = setTimeout(function() {
            let bounds = window.getBounds()
            bounds.x = bounds.x - boundsCorrection.x
            bounds.y = bounds.y - boundsCorrection.y
            bounds.width = bounds.width - boundsCorrection.width
            bounds.height = bounds.height - boundsCorrection.height

            console.log("saving: ", bounds)
            saveSettings(bounds)
        }, 100)
    }

    let settings = {}
    try {
        settings = JSON.parse(fs.readFileSync(settingsPath))
    }
    catch (err) {}

    console.log("loaded settings:", settings)

    settings = settings || {
        x : 0,
        y : 0,
        width : 300,
        height : 300
    }

    let window = new BrowserWindow(settings)
    let bounds = window.getBounds();
    let boundsCorrection = {
        x: bounds.x - settings.x,
        y: bounds.y - settings.y,
        width : bounds.width - settings.width,
        height : bounds.height - settings.height
    }
    console.log("window bounds: ", window.getBounds())
    console.log("correction: ", boundsCorrection)
    window.setMenu(null);
	/*console.log("setting bounds:", settings)
	window.setBounds(settings);
    let actualBounds = window.getBounds();
    let correctedBounds = {
        x : settings.x - (actualBounds.x - settings.x),
        y : settings.y - (actualBounds.y - settings.y),
        width : settings.width - (actualBounds.width - settings.width),
        height : settings.height - (actualBounds.height - settings.height),
    }
    console.log("setting corrected bounds: ", correctedBounds)
    window.setBounds(correctedBounds)
    console.log("window bounds: ", window.getBounds())

    //window.webContents.openDevTools()*/
    window.loadURL(`file://${__dirname}/${appName}/${appName}.html`)
    window.on("resize", boundsChanged);
    window.on("move", boundsChanged);
    window.on("ready-to-show", () => console.log("ready to show"))
    return window
}


let activateApp = function(appName) {
    if (windows[appName]) {
        windows[appName].show();
        windows[appName].focus();
    }
    else {
        windows[appName] = createWin(appName)
        windows[appName].on('closed', function () {
            console.log("Deleting..")
            delete windows[appName]
        })

    }
}

let respond = function(response, code) {
    response.writeHead(code)
    response.end()
}

let handleRequest = function(request, response) {
    try {
        console.log(request.method, " mod: ", request.url);
        if (request.method === "POST") {
            let action = request.url.slice(1) // remove starting '/'
            if (action === 'quit') {
                respond(response, 200)
                app.quit()
            }
            else if (["do"].includes(action)) {
                respond(response, activateApp(action) ? 200 : 500)
            }
            else {
                respond(response, 404)
            }
        }
        else if (request.method === "GET") {
            respond(response, 200)
        }
        else {
            respond(response, 405)
        }
    }
    catch (err) {
        console.log(err);
    }
}
let http = require('http')

let server = http.createServer(handleRequest)

try {
	fs.unlink("/run/user/1000/org.restfulipc.refude.desktop")
}
catch (err) {
}

server.listen("/run/user/1000/org.restfulipc.refude.desktop")


