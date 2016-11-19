
const {app, BrowserWindow} = require('electron')
const path = require('path')
const url = require('url')
const windowManager = require('../common-js/createwin.js')


let win

const shouldQuit = app.makeSingleInstance((commandLine, workingDirectory) => {
  // Someone tried to run a second instance, we should focus our window.
  if (win) {
	win.restore()
	win.show()
    win.focus()
  }
})

if (shouldQuit) {
  app.quit()
}


app.on('ready', () => {
 	win = windowManager.createWin("appconfig")
	win.on("closed", () => { win = null; });
    win.setMenu(null);
	win.loadURL(`file://${__dirname}/appconfig.html`);
	//win.webContents.openDevTools();
});


// Quit when all windows are closed.
app.on('window-all-closed', () => {
    app.quit();
});

