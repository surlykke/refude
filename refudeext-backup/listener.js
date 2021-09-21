const refudeUrl = 'http://localhost:7938/refude/'


let refudeTabs = handler => {
    chrome.tabs.query({ url: "http://localhost/*" }, tabs => {
        handler(tabs.filter(t => t.url.startsWith(refudeUrl))) 
    })
}

chrome.runtime.onInstalled.addListener(() => {
    watchSse()
});

chrome.runtime.onMessage.addListener(request => { 
    if (request === "dismiss") { 
        refudeTabs(tabs => tabs.forEach(tab => chrome.tabs.remove(tab.id))) 
    }
})

let watchSse = () => {
    let evtSource = new EventSource("http://localhost:7938/watch")

    evtSource.onopen = this.getResource

    evtSource.onerror = event => {
        if (evtSource.readyState === 2) {
            setTimeout(watchSse, 5000)
        }
    }

    evtSource.onmessage = ({ data }) => {
        if (data === '/client') {
            refudeTabs(tabs => {
                if (tabs[0]) {
                    chrome.windows.get(tabs[0].windowId, window => {
                        chrome.windows.update(window.id, { focused: true })
                        chrome.tabs.update(tabs[0].id, { active: true })
                    })
                } else {
                    chrome.tabs.create({ url: refudeUrl })
                }
            })
        }
    }
}