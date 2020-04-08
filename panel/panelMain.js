// panel specific stuff to be executed in main process.

let {BrowserWindow, ipcMain, app} = require('electron')
let {loadBounds, saveBounds, onDisplayChange} = require('../common/managewindows')

let path = require('path')
let url = require('url')

let panelWindow


let createPanel = () => {
    panelWindow = new BrowserWindow({show: false, frame: false, alwaysOnTop: true, webPreferences: { nodeIntegration: true } })

    panelWindow.loadURL(url.format({
        pathname: path.join(__dirname, '/panel.html'),
        protocol: 'file:',
        slashes: true
    }))
    panelWindow.once('ready-to-show', () => {
        let {x, y} = Object.assign({x: 100, y: 50, width: 280, height: 32}, loadBounds("panel"))
        panelWindow.setPosition(x, y)
        panelWindow.show()
    })
    panelWindow.on('move', () => saveBounds("panel", panelWindow.getBounds()))
    panelWindow.on('closed', app.quit)
//    app.isPackaged || panelWindow.webContents.openDevTools()
}

let adjustWindowSize = (evt, rect) => {
    let {width, height} = rect
    panelWindow.setSize(width + 5, height + 1)
}

let panelBounds = () => {
    return panelWindow.getBounds()
}

ipcMain.on('panelClose', () => app.quit())
ipcMain.on('panelSizeChange', adjustWindowSize)

module.exports = {createPanel, panelBounds}