
const {app, BrowserWindow} = require('electron')
const path = require('path')
const url = require('url')
const windowManager = require('../common-js/createwin.js')


// Keep a global reference of the window object, if you don't, the window will
// be closed automatically when the JavaScript object is garbage collected.
let win

app.on('ready', () => {
 	win = windowManager.createWin("panel", {transparent: true, frame: false})
	win.on("closed", () => { win = null; });
    win.setMenu(null);
	win.setAlwaysOnTop(true);    
	win.loadURL(`file://${__dirname}/panel.html`);
	//win.webContents.openDevTools();
});


// Quit when all windows are closed.
app.on('window-all-closed', () => {
    app.quit();
});

