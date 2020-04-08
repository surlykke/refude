process.env.ELECTRON_DISABLE_SECURITY_WARNINGS = true

const  {app} = require('electron')
const {initializeWindowManager} = require('./common/managewindows')

let {createPanel} = require('./panel/panelMain')
let {createDoWindow} = require('./do/doMain')
let {createIndicatorWindow} = require('./indicator/indicatorMain')

app.on('ready', () => {
    /*app.isPackaged || BrowserWindow.addDevToolsExtension(
        '/home/surlykke/snap/chromium/1077/.config/chromium/Default/Extensions/fmkadmapgofadopljbjfkapdkoienihi/4.6.0_0'
    )*/

    initializeWindowManager()
    createPanel()
    createDoWindow()
    createIndicatorWindow()
})
