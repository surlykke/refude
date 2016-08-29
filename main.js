const electron = require('electron')
const app = electron.app

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

let activateApp = function(appName) {
    if (windows[appName]) {
        windows[appName].show();
        windows[appName].focus();
    }
    else {
        console.log("Creating", appName)
        windows[appName] = new BrowserWindow({width: 600, height: 400})
        //windows[appName].webContents.openDevTools();     
        windows[appName].setMenu(null);
        windows[appName].loadURL(`file://${__dirname}/${appName}/${appName}.html`)
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

fs.unlink("/run/user/1000/org.restfulipc.refude.desktop")
server.listen("/run/user/1000/org.restfulipc.refude.desktop")


