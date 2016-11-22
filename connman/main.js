const {app} = require('electron')

const singletonMaker = require('../common-js/singletonapp.js')

app.on('ready', () => {
	singletonMaker.singletonApp("/run/user/1000/org.refude.apps.connman", __dirname);
});

