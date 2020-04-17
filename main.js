process.env.ELECTRON_DISABLE_SECURITY_WARNINGS = true

const {app, BrowserWindow, ipcMain, screen} = require('electron')
let { manageWindow } = require('./common/managewindows')
let url = require('url')
let path = require('path')
let http = require('http')

let panelWindow

let createPanel = () => {
    panelWindow = new BrowserWindow({show: false, frame: false, alwaysOnTop: true, webPreferences: { nodeIntegration: true } })

    panelWindow.loadURL(url.format({
        pathname: path.join(__dirname, '/panel/panel.html'),
        protocol: 'file:',
        slashes: true
    }))
    panelWindow.once('ready-to-show', () => {
        manageWindow(panelWindow, 'panel', true, false)
        panelWindow.show()
        ipcMain.on('panelSizeChange', (evt, rect) => panelWindow.setSize(rect.width + 5, rect.height + 1))
    })
    panelWindow.on('closed', app.quit)
    //app.isPackaged || panelWindow.webContents.openDevTools()
}

let doWindow
let server

let createDoWindow = () => {
    doWindow = new BrowserWindow({
        x: 100, y: 100, width: 300, height: 500,
        show: false, frame: false, alwaysOnTop: true, webPreferences: { nodeIntegration: true }
    })

    doWindow.loadURL(url.format({
        pathname: path.join(__dirname, '/do/do.html'),
        protocol: 'file:',
        slashes: true
    })).then(() => {
        //doWindow.webContents.openDevTools()
        manageWindow(doWindow, "do", false, true)
        doWindow.on('closed', () => { win = undefined })

        server = http.createServer(function (req, res) {
            res.end('')

            if (!doWindow.isVisible()) {
                let pb = panelWindow.getBounds()
                doWindow.setPosition(pb.x, pb.y + pb.height + 12)
                doWindow.show()
                doWindow.webContents.send("doReset")
                indicatorWindow.send("screens", screen.getAllDisplays())
            } else {
                doWindow.send("doMove", req.url !== "/up")
            }

        }).listen("/run/user/1000/org.refude.panel.do");

        ipcMain.on("doResourceSelected", (evt, res) => {
            if (doWindow.isVisible() && res && res.Type === "window") { 
                indicatorWindow.showInactive()
                indicatorWindow.webContents.send("resource", res)
            } else {
                indicatorWindow.hide()
            }
        })

        ipcMain.on("doClose", () => {
            doWindow.hide()
            indicatorWindow.hide()
        })
    })
}

let indicatorWindow

let createIndicatorWindow = () => {
    indicatorWindow = new BrowserWindow({
        show: false, frame: false, transparent: true, skipTaskbar: true, webPreferences: { nodeIntegration: true }
    })

    indicatorWindow.loadURL(url.format({
        pathname: path.join(__dirname, '/indicator/indicator.html'),
        protocol: 'file:',
        slashes: true
    })).then(() => {
        manageWindow(indicatorWindow, "indicator", true, true)
    }).catch(error => console.log("Error in load:", error))

    //indicatorWindow.webContents.openDevTools()
}


app.on('ready', () => {
    /*app.isPackaged || BrowserWindow.addDevToolsExtension(
        '/home/surlykke/snap/chromium/1077/.config/chromium/Default/Extensions/fmkadmapgofadopljbjfkapdkoienihi/4.6.0_0'
    )*/

    createPanel()
    createDoWindow()
    createIndicatorWindow()

    ipcMain.on('panelClose', () => app.quit())

})
