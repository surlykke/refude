const refudeUrl = 'http://localhost:7938/refude/'

let currentWindowId 

let windowOpts
let x, y, width, height
let xCorr, yCorr 

chrome.windows.onBoundsChanged.addListener(w => {
    if (w.id === currentWindowId) {
        x = w.left
        y = w.top
        width = w.width
        height = w.height 
        if (windowOpts?.left && windowOpts?.top) {
            xCorr = x - windowOpts.left
            yCorr = y - windowOpts.top
            windowOpts = undefined
        }
    }
})



chrome.runtime.onMessage.addListener(request => {
    if (request === "dismiss") {
        chrome.tabs.query({}, tabs => tabs.filter(t => t.url.startsWith(refudeUrl)).forEach(t => chrome.tabs.remove(t.id)))
        if (x && y && width && height) {
            if (xCorr !== undefined) x = x - xCorr
            if (yCorr !== undefined) y = y - yCorr
            chrome.storage.local.set({x: x, y: y, width: width, height: height})
        }
        currentWindowId = x = y = width = height = undefined
    } 
})



let handleActivation = (tabs, up) => {
    for (let tab of tabs.reverse()) { // In case of more than one tab open we
                                      // want to focus the latest.
        if (tab.url.startsWith(refudeUrl)) {
            chrome.windows.get(tab.windowId, window => {
                currentWindowId = window.id
                if (tab.active && window.focused) {
                    chrome.tabs.sendMessage(tab.id, up ? "moveUp" : "moveDown")
                } else {
                    chrome.windows.update(window.id, { focused: true })
                    chrome.tabs.update(tab.id, { active: true })
                }
            })
            return
        }

    }
    // None found, so
    chrome.storage.local.get(['x', 'y', 'width', 'height'], result => {
        windowOpts = { url: refudeUrl, type: "popup"}
        windowOpts.left = result.x
        windowOpts.top = result.y
        windowOpts.width = result.width
        windowOpts.height = result.height
        chrome.windows.create(windowOpts, window => {
            currentWindowId = window.id
        })
    });
}

let socket 

let listen = () => {
    try {
        socket = new WebSocket("ws://localhost:7938/client/control")
    } catch (err) {
        socket = undefined
        setTimeout(listen, 1000)
    }
    socket.onclose = ev => setTimeout(listen, 1000)
    socket.onmessage = ({data}) => {
        chrome.tabs.query({}, tabs => handleActivation(tabs, data === "up"))
    }
}
listen()


