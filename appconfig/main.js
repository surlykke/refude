const {app} = require('electron')

const singletonMaker = require('../common/singletonapp.js')

app.on('ready', () => {
	singletonMaker.singletonApp(process.env.XDG_RUNTIME_DIR + "/org.refude.apps.appconfig", __dirname);
});

