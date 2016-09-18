const electron = require('electron');
const app = electron.app;
const BrowserWindow = electron.BrowserWindow;
const fs = require('fs');

exports.createWin = function(appName) {
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
            saveBounds({ x: bounds.x - boundsCorrection.x, 
                         y: bounds.y - boundsCorrection.y, 
                         width: bounds.width - boundsCorrection.width, 
                         height: bounds.height - boundsCorrection.height 
                     });
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

    let loadedWinOptions;
    try {
        loadedWinOptions = JSON.parse(fs.readFileSync(`${__dirname}/${appName}/windowoptions.json`));
    }
    catch (err) {
        loadedWinOptions = {};
    }
    let opts = Object.assign({}, loadedBounds, loadedWinOptions);
    opts.minWidth = 30;
    opts.minHeight = 20;
    let window = new BrowserWindow(opts);
    
    //opts.alwaysOnTop && window.setAlwaysOnTop(true);

	//window.webContents.openDevTools({detach: true});
    let actualBounds = window.getBounds();
    let boundsCorrection = {
        x: actualBounds.x - loadedBounds.x,
        y: actualBounds.y - loadedBounds.y,
        width : actualBounds.width - loadedBounds.width,
        height : actualBounds.height - loadedBounds.height
    };
   
    window.setMenu(null);
    window.loadURL(`file://${__dirname}/${appName}/${appName}.html`);
    window.on("resize", boundsChanged);
    window.on("move", boundsChanged);

    // band aid
    opts.alwaysOnTop && window.setAlwaysOnTop(true);
    return window;
};

