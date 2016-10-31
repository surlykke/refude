const electron = require('electron')
const app = electron.app
const fs = require('fs')
const windowManager = require('./createwin.js')
const log = console.log;

app.on('window-all-closed', function () {})
app.log = console.log;

let windows = {}

let activateApp = function (appName) {
    console.log("appName:", appName)
    if (windows[appName]) {
        windows[appName].show();
        windows[appName].focus();
    } else {
        windows[appName] = windowManager.createWin(appName)
        windows[appName].on('closed', function () {
            console.log("Deleting..")
            delete windows[appName]
        })

    }
}

let handleRequest = function (request, response) {

    let respond = function (response, code) {
        response.writeHead(code)
        response.end()
    }

    try {
        console.log(request.method, " mod: ", request.url);
        if (request.method === "POST") {
            let action = request.url.slice(1) // remove starting '/'
            if (action === 'quit') {
                respond(response, 200)
                app.quit()
            } else if (["do", "appconfig", "panel", "power"].includes(action)) {
                console.log("do")
                respond(response, activateApp(action) ? 200 : 500)
            } else {
                respond(response, 404)
            }
        } else if (request.method === "GET") {
            respond(response, 200)
        } else {
            respond(response, 405)
        }
    } catch (err) {
        console.log(err);
    }
}

app.on('ready', function () {
    let http = require('http')
    let server = http.createServer(handleRequest)

    http.get({"socketPath": "/run/user/1000/org.restfulipc.refude.desktop"}, (res) => {
        res.resume();
        console.log("Refude seems to be already running. Exitting this instance...")
        process.exit(0)
    }).on("error", (err) => {
        try {
            fs.unlinkSync("/run/user/1000/org.restfulipc.refude.desktop")
        }
        catch (err) {
        }

        server.listen("/run/user/1000/org.restfulipc.refude.desktop")
    });
});

