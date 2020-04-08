// Do specific stuff to be executed in main process.

let { BrowserWindow, ipcMain } = require('electron')
let { saveBounds, loadBounds } = require('../common/managewindows')
let { panelBounds } = require('../panel/panelMain')
let { indicatorWindowPrepareToShow, indicatorWindowShow } = require('../indicator/indicatorMain')

let path = require('path')
let url = require('url')
let http = require('http')

let server

let doWindow

let createDoWindow = () => {
    doWindow = new BrowserWindow({
        x: 100, y: 100, width: 300, height: 500,
        show: false, frame: false, alwaysOnTop: true, webPreferences: { nodeIntegration: true }
    })

    doWindow.loadURL(url.format({
        pathname: path.join(__dirname, '/do.html'),
        protocol: 'file:',
        slashes: true
    })).then(() => {
        //doWindow.webContents.openDevTools()
        doWindow.on('resize', () => saveBounds("do", doWindow.getBounds()))
        doWindow.on('closed', () => { win = undefined })

        server = http.createServer(function (req, res) {
            res.end('')

            if (!doWindow.isVisible()) {
                let pb = panelBounds()
                let bounds = loadBounds('do') || { width: pb.width + 40, height: 500 }
                doWindow.setBounds({ x: pb.x, y: pb.y + pb.height + 12, width: bounds.width, height: bounds.height })
                doWindow.show()
                doWindow.webContents.send("doReset")
                indicatorWindowPrepareToShow()
            } else {
                doWindow.send("doMove", req.url !== "/up")
            }

        }).listen("/run/user/1000/org.refude.panel.do");

        ipcMain.on("doResourceSelected", (evt, res) => {
            doWindow.isVisible() && indicatorWindowShow(res)
        })

        ipcMain.on("doClose", () => {
            doWindow.hide()
            indicatorWindowShow(null)
        })
    })
}


module.exports = { createDoWindow }