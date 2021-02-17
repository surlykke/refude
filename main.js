process.env.ELECTRON_DISABLE_SECURITY_WARNINGS = true

const { app, BrowserWindow, ipcMain } = require('electron')
let { rememberBounds, manageWindow } = require('./common/managewindows')
let url = require('url')
let path = require('path')
let http = require('http')


let panelWindow
let server

let createPanel = () => {
    panelWindow = new BrowserWindow({ show: false, frame: false, alwaysOnTop: true, webPreferences: { nodeIntegration: true, enableRemoteModule: true } })

    panelWindow.loadURL(url.format({
        pathname: path.join(__dirname, '/panel/panel.html'),
        protocol: 'file:',
        slashes: true
    }))
    panelWindow.once('ready-to-show', () => {
        manageWindow(panelWindow, 'panel')
        panelWindow.showInactive()
        ipcMain.on('panelSizeChange', (evt, width, height) => {
            let scaledWidth = Math.round(panelWindow.webContents.zoomFactor * width)
            let scaledHeight = Math.round(panelWindow.webContents.zoomFactor * height)
            panelWindow.setSize(scaledWidth, scaledHeight)
        })

        ipcMain.on('devtools', () => {
            panelWindow.webContents.openDevTools()
        })

        server = http.createServer(function (req, res) {
            res.end('')
            rememberBounds('panel', panelWindow.getBounds())
            panelWindow.send("show", req.url === "/up")
            panelWindow.focus()
        }).listen("/run/user/1000/org.refude.panel.do");

    })


    panelWindow.on('closed', app.quit)
}


app.on('ready', () => {
    createPanel()
    ipcMain.on('panelClose', () => {
        app.exit(0)
    })
    ipcMain.on('panelMinimize', () => {
        panelWindow.hide()
        setTimeout(() => panelWindow.showInactive(), 5000)
    })
})
