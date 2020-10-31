process.env.ELECTRON_DISABLE_SECURITY_WARNINGS = true

const { app, BrowserWindow, ipcMain, screen } = require('electron')
let { rememberBounds, manageWindow } = require('./common/managewindows')
let url = require('url')
let path = require('path')
let http = require('http')


let setupWatch = () => {
    let windows = [panelWindow, doWindow, osdWindow, indicatorWindow]
    ipcMain.on("sseopen", () => {
        windows.forEach(w => w.webContents.send("sseopen"))
    })
    
    ipcMain.on("sseerror", () => {
        windows.forEach(w => w.webContents.send("sseerror"))
    })
    
    ipcMain.on("ssemessage", (event, path) => {
        windows.forEach(w => w.webContents.send(path))
    })

}

let panelWindow

let createPanel = () => {
    panelWindow = new BrowserWindow({ show: false, frame: false, alwaysOnTop: true, webPreferences: { nodeIntegration: true } })

    panelWindow.loadURL(url.format({
        pathname: path.join(__dirname, '/panel/panel.html'),
        protocol: 'file:',
        slashes: true
    }))
    panelWindow.once('ready-to-show', () => {
        manageWindow(panelWindow, 'panel')
        panelWindow.show()
        ipcMain.on('panelSizeChange', (evt, width, height) => {
            let scaledWidth = Math.round(panelWindow.webContents.zoomFactor * width)
            let scaledHeight = Math.round(panelWindow.webContents.zoomFactor * height)
            panelWindow.setSize(scaledWidth, scaledHeight)
        })

    })
    panelWindow.on('closed', app.quit)
    // panelWindow.webContents.openDevTools()
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
        
       server = http.createServer(function (req, res) {
            res.end('')

            if (!doWindow.isVisible()) {
                doWindow.show()
                doWindow.webContents.send("doShow")
                indicatorWindow.send("screens", screen.getAllDisplays())
            } else {
                doWindow.send("doMove", req.url === "/up")
                doWindow.focus()
            }

        }).listen("/run/user/1000/org.refude.panel.do");

 
        manageWindow(doWindow, "do")
        doWindow.on('closed', () => { win = undefined })
        
        ipcMain.on("doLinkSelected", (evt, link) => {
            if (doWindow.isVisible() && link && link.profile === "/profile/window" && (!link.meta || link.meta["state"] !== "minimized")) {
                indicatorWindow.showInactive()
                indicatorWindow.webContents.send("linkSelected", link)
            } else {
                indicatorWindow.hide()
            }
        })

        
        ipcMain.on("doClose", () => {
            rememberBounds('panel', panelWindow.getBounds())
            rememberBounds('do', doWindow.getBounds())
            rememberBounds('indicator', indicatorWindow.getBounds())
            doWindow.hide()
            indicatorWindow.hide()
        })
    })
        
    //doWindow.webContents.openDevTools()
}

let indicatorWindow

let createIndicatorWindow = () => {
    indicatorWindow = new BrowserWindow({
        show: false, frame: false, skipTaskbar: true, transparent: true, webPreferences: { nodeIntegration: true }
    })

    indicatorWindow.loadURL(url.format({
        pathname: path.join(__dirname, '/indicator/indicator.html'),
        protocol: 'file:',
        slashes: true
    })).then(() => {
        manageWindow(indicatorWindow, "indicator")
    }).catch(error => console.error(error))

    //indicatorWindow.webContents.openDevTools()
}

let osdWindow

let createOsdWindow = () => {
    osdWindow = new BrowserWindow({
        show: false, frame: false, transparent: true, alwaysOnTop: true, webPreferences: { nodeIntegration: true }
    })

    osdWindow.loadURL(url.format({
        pathname: path.join(__dirname, "/osd/osd.html"),
        protocol: 'file',
        slashes: true,
    })).then(() => {
        ipcMain.on('osdHide', () => osdWindow.hide())
        ipcMain.on('osdShow', (evt, rect) => {
            let pb = panelWindow.getBounds()
            let zf = panelWindow.webContents.zoomFactor
            let [width, height] = [Math.round(zf*rect.width), Math.round(zf*rect.height)]
            width = Math.max(width, pb.width)
            osdWindow.setBounds({ x: pb.x, y: pb.y + pb.height + 12, width: width, height: height})
            osdWindow.webContents.zoomFactor = zf
            osdWindow.showInactive()

        })
    })

    //osdWindow.webContents.openDevTools()
}

app.on('ready', () => {
    createPanel()
    createDoWindow()
    createIndicatorWindow()
    createOsdWindow()
    setupWatch()

    ipcMain.on('panelClose', () => app.quit())

})
