let { BrowserWindow, screen } = require('electron')
let { loadBounds, saveBounds } = require('../common/managewindows')
let url = require('url')
let path = require('path')

let indicatorWindow

let createIndicatorWindow = () => {
    indicatorWindow = new BrowserWindow({
        show: false, frame: false, transparent: true, skipTaskbar: true, webPreferences: { nodeIntegration: true }
    })

    indicatorWindow.loadURL(url.format({
        pathname: path.join(__dirname, '/indicator.html'),
        protocol: 'file:',
        slashes: true
    })).then(() => {
        let bounds = Object.assign({ x: 100, y: 100, width: 400, height: 300 }, loadBounds("indicator"))
        indicatorWindow.setBounds(bounds)
        indicatorWindow.on("resize", () => saveBounds("indicator", indicatorWindow.getBounds()))
        indicatorWindow.on("move", () => saveBounds("indicator", indicatorWindow.getBounds()))
    }).catch(error => console.log("Error in load:", error))

    //indicatorWindow.webContents.openDevTools()
}

let indicatorWindowPrepareToShow = () => {
    indicatorWindow.webContents.send("screens", screen.getAllDisplays())
    indicatorWindow.send("screens", screen.getAllDisplays())
}

let indicatorWindowShow = (resource) => {
    if (resource && resource.Type === "window") {
        indicatorWindow.showInactive()
    } else {
        indicatorWindow.hide()
    }
    
    indicatorWindow.webContents.send("resource", resource)
}

module.exports = { createIndicatorWindow, indicatorWindowPrepareToShow, indicatorWindowShow }



