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
let indicatorDismissedInThisShowing = false

let createDoWindow = () => {
    doWindow = new BrowserWindow({
        x: 100, y: 100, width: 300, height: 500,
        show: false, frame: false, alwaysOnTop: true, webPreferences: { nodeIntegration: true }
    })
    doWindow.removeMenu()

    doWindow.loadURL(url.format({
        pathname: path.join(__dirname, '/do/do.html'),
        protocol: 'file:',
        slashes: true
    })).then(() => {
        
       server = http.createServer(function (req, res) {
            res.end('')
            
            rememberBounds('panel', panelWindow.getBounds())

            if (!doWindow.isVisible()) {
                doWindow.show()
                indicatorDismissedInThisShowing = false 
                doWindow.webContents.send("doShow")
                indicatorWindow.send("screens", screen.getAllDisplays())
            } else {
                doWindow.send("doMove", req.url === "/up")
                doWindow.focus()
            }

        }).listen("/run/user/1000/org.refude.panel.do");

 
        manageWindow(doWindow, "do")
        
        ipcMain.on("doLinkSelected", (evt, link) => {
            if (!indicatorDismissedInThisShowing && 
                doWindow.isVisible() && 
                link && 
                link.profile === "/profile/window" && 
                (!link.meta || link.meta["state"] !== "minimized")) {

                indicatorWindow.showInactive()
                indicatorWindow.webContents.send("linkSelected", link)
            } else {
                indicatorWindow.hide()
            }
        })

        doWindow.on('close', e => {
            e.preventDefault()
            dismissDo()
        })

        ipcMain.on("dismiss", dismissDo)
    })
        
    //doWindow.webContents.openDevTools()
}

let dismissDo = () => {
    if (doWindow) {
        rememberBounds('do', doWindow.getBounds())
        doWindow.hide()
    }

    if (indicatorWindow) {
        rememberBounds('indicator', indicatorWindow.getBounds())
        indicatorWindow.hide()
    }
}


let indicatorWindow

let createIndicatorWindow = () => {
    indicatorWindow = new BrowserWindow({
        show: false, frame: false, skipTaskbar: true, webPreferences: { nodeIntegration: true }
    })

    indicatorWindow.removeMenu()

    indicatorWindow.on('close', e => {
        e.preventDefault()
        indicatorDismissedInThisShowing = true
        indicatorWindow.hide()
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

    ipcMain.on('panelClose', () => {
        app.exit(0)
    })

})
