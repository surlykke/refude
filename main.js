process.env.ELECTRON_DISABLE_SECURITY_WARNINGS = true

const { app, BrowserWindow, ipcMain, screen } = require('electron')
let { manageWindow } = require('./common/managewindows')
let url = require('url')
let path = require('path')
let http = require('http')

let panelWindow

let createPanel = () => {
    panelWindow = new BrowserWindow({ show: false, frame: false, alwaysOnTop: true, webPreferences: { nodeIntegration: true } })

    panelWindow.loadURL(url.format({
        pathname: path.join(__dirname, '/panel/panel.html'),
        protocol: 'file:',
        slashes: true
    }))
    panelWindow.once('ready-to-show', () => {
        manageWindow(panelWindow, 'panel', true, false)
        panelWindow.show()
        ipcMain.on('panelSizeChange', (evt, rect) => {
            let width = Math.round(panelWindow.webContents.zoomFactor*rect.width) + 8 
            let height = Math.round(panelWindow.webContents.zoomFactor*rect.height) + 1
            panelWindow.setSize(width, height)
        })

    })
    panelWindow.on('closed', app.quit)
    //panelWindow.webContents.openDevTools()
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
    }).catch(error => console.error(error))

}

let osdWindow

let createOsdWindow = () => {
    osdWindow = new BrowserWindow({
        show: false, frame: false, transparent: true, skipTaskbar: true, webPreferences: { nodeIntegration: true }
    })

    osdWindow.loadURL(url.format({
        pathname: path.join(__dirname, "/osd/osd.html"),
        protocol: 'file',
        slashes: true,
    })).then(() => {
        ipcMain.on('osdShow', (evt, shown) => {
            if (shown) {
                osdWindow.webContents.setZoomFactor(panelWindow.webContents.zoomFactor)
                osdWindow.showInactive()
            } else {
               osdWindow.hide()
            }
        })
        ipcMain.on('osdSize', (evt, rect) => {
            let pb = panelWindow.getBounds()
            let {x, y, width} = {x: pb.x, y: pb.y + pb.height + 12, width: pb.width}
            let height = Math.round(osdWindow.webContents.zoomFactor*rect.height) + 1
            osdWindow.setBounds({ x: x, y: y, width: width, height: height})
        })
    })

    // osdWindow.webContents.openDevTools()
}

app.on('ready', () => {

    createPanel()
    createDoWindow()
    createIndicatorWindow()
    createOsdWindow()

    ipcMain.on('panelClose', () => app.quit())

})
