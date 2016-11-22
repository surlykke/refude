const electron = require('electron');
const {app, BrowserWindow} = electron;
const fs = require('fs');

exports.createWin = function(appName, windowOpts) {
    let boundsPath = app.getPath('userData') + "/" + appName + "-bounds.json" 
   
    let saveBounds = function(bounds) {
        fd = fs.openSync(boundsPath, "w");
        fs.writeFileSync(fd, JSON.stringify(bounds));
    };

    let timeoutId = null;
    let boundsChanged = function() {
        if (timeoutId) {
            clearTimeout(timeoutId);
        }
        timeoutId = setTimeout(function() {
            let bounds = window.getBounds();
            console.log("bounds", bounds);
            saveBounds(bounds);
            timeoutId = null;
        }, 100);
    };

    let loadedBounds;
    try {
        loadedBounds = JSON.parse(fs.readFileSync(boundsPath));
    }
    catch (err) {
        loadedBounds = {x : 0, y : 0, width : 300, height : 300 };
    }

	windowOpts = windowOpts || {};
	Object.assign(windowOpts, loadedBounds);
	let window =  new BrowserWindow(windowOpts);
    window.on("resize", boundsChanged);
    window.on("move", boundsChanged);
	if (windowOpts.alwaysOnTop) {
		window.setAlwaysOnTop(true);
	}
	if (windowOpts.openDevTools) {
		win.webContents.openDevTools();
	}
    return window;
};

