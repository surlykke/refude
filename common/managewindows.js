// Tools for maintaining position and size for windows
// To be executed in main

const { screen, app } = require('electron')
const fs = require('fs')
const dataPath = app.getPath('userData') + "/windowData.json"

let windowData
let signature

let signatureChangeListeners = []

let onDisplayChange = () => signatureChangeListeners.forEach(scl => scl())

let getSignature = () => {
    let displays = screen.getAllDisplays()
    displays = displays.sort((d1, d2) => d1.id - d2.id)
    signature = displays.map(d => d.id + "-" + d.bounds.x + "-" + d.bounds.y + "-" + d.bounds.width + "-" + d.bounds.height).join("-")
    windowData[signature] = windowData[signature] || {}
    signatureChangeListeners.forEach(scl => scl())
}


// Must happen before first call to manageWindow
let initializeWindowManager = () => {
    try {
        windowData = JSON.parse(fs.readFileSync(dataPath))
    } catch(error) {}

    if ('object' !== typeof windowData || Array.isArray(windowData)) {
        windowData = {}
    }
    
    screen.on('display-metrics-changed', getSignature)
    getSignature() 
}

let persistScheduled
let persist = () => {
    if (!persistScheduled) {
        persistScheduled = true;
        setTimeout(() => { 
            try {
                fs.writeFileSync(dataPath, JSON.stringify(windowData)); persistScheduled = undefined 
            } catch(error) {
                console.log("Unable to save", dataPath, error)
            }
            }, 2000)
    }
}

let saveBounds = (windowName, bounds) => {
    windowData[signature][windowName] = bounds
    persist()
}

let loadBounds = (windowName, bounds) => {
    return windowData[signature][windowName]
}


module.exports = { saveBounds, loadBounds, onDisplayChange, initializeWindowManager}