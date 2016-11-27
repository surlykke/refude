const {webFrame} = require('electron');
const zoomLevelKey = document.location.href + "@@zoomLevel";

let setZoomLevel = function(level) {
    console.log("Setting zoomlevel to", level);
    webFrame.setZoomLevel(level);
    zoomLevel = webFrame.getZoomLevel();
    console.log("Saving zoomLevel:", zoomLevel);
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
console.log("Loaded zoomlevel:", zoomLevel);
document.addEventListener("keydown", keyDown);
