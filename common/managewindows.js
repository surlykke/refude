// Tools for maintaining position and size for windows
// To be executed in main

const { screen, app } = require('electron')
const fs = require('fs')
const dataPath = app.getPath('userData') + "/windowData.json"

let windowData

let initialized
let initialize = () => {
    if (!initialized) {
        try {
            windowData = JSON.parse(fs.readFileSync(dataPath))
        } catch (error) { }

        if ('object' !== typeof windowData || Array.isArray(windowData)) {
            windowData = {}
        }
        initialized = true
    }
}


let getSignature = () => {
    let displays = screen.getAllDisplays()
    displays = displays.sort((d1, d2) => d1.id - d2.id)
    let signature = displays.map(d => d.id + "-" + d.bounds.x + "-" + d.bounds.y + "-" + d.bounds.width + "-" + d.bounds.height).join("-")
    windowData[signature] = windowData[signature] || {}
    return signature
}


let persistScheduled
let persist = () => {
    if (!persistScheduled) {
        persistScheduled = true;
        setTimeout(() => {
            try {
                fs.writeFileSync(dataPath, JSON.stringify(windowData)); persistScheduled = undefined
            } catch (error) {
                console.error("Unable to save", dataPath, error)
            }
        }, 2000)
    }
}

let saveBounds = (windowName, bounds) => {
    let signature = getSignature()
    windowData[signature][windowName] = bounds
    persist()
}

let loadBounds = (windowName) => {
    let signature = getSignature()
    return windowData[signature][windowName]
}

let manageWindow = (window, windowName, managePosition, manageSize) => {
    initialize()
    let set = () => {
        let bounds = loadBounds(windowName)
        if (bounds) {
            managePosition && window.setPosition(bounds.x, bounds.y)
            manageSize && window.setSize(bounds.width, bounds.height)
        }
    }
    set()
    screen.on('display-metrics-changed', set)
    window.on('move', () => saveBounds(windowName, window.getBounds()))
    window.on('resize', () => saveBounds(windowName, window.getBounds()))
}

module.exports = {manageWindow} 