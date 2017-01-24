const {webFrame} = require('electron');
const zoomLevelKey = document.location.href + "@@zoomLevel";

let setZoomLevel = function(level) {
    webFrame.setZoomLevel(level);
    zoomLevel = webFrame.getZoomLevel();
    localStorage.setItem(zoomLevelKey, zoomLevel);
}

let keyDown = function(event) {
    if ("+" === event.key && event.altKey) {
        setZoomLevel(zoomLevel + 1);
    }
    else if ("-" === event.key && event.altKey) {
        setZoomLevel(zoomLevel - 1);
    }
};

let zoomLevel = parseInt(localStorage.getItem(zoomLevelKey));
if (isNaN(zoomLevel)) {
    zoomLevel = 0;
}
setZoomLevel(zoomLevel);
document.addEventListener("keydown", keyDown);
