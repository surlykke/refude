let tabsSocket
let commandSocket

const reportTabs = () => {
    chrome.tabs.query({}, tabs => {
        if (tabsSocket) {
            let tabsData = tabs.map(t => {
                return {
                    id: "" + t.id,
                    title: t.title,
                    url: t.url
                }
            })
            tabsSocket.send(JSON.stringify(tabsData))
        }
    })
}

const openTabsSocket = () => {
    tabsSocket = new WebSocket("ws://localhost:7938/tab/websocket")
    tabsSocket.onopen = reportTabs
    tabsSocket.onmessage = m => {
        let marshalled = JSON.parse(m.data)
        let tabId = parseInt(marshalled)
        if (tabId) {
            console.log("message, tabId:", tabId)
            var updateProperties = { 'active': true };
            chrome.tabs.update(tabId, updateProperties, (tab) => { })
        }
    }
    tabsSocket.onclose = () => {
        tabsSocket = null;
        setTimeout(openTabsSocket, 10000)
    }
}

const openCommandSocket = () => {
    console.log("Open command socket")
    commandSocket = new WebSocket("ws://localhost:7938/refude/commands")
    commandSocket.onmessage = m => {
        console.log("commandSocket data:", m.data)
        let marshalled = JSON.parse(m.data)
        switch (marshalled.Command) {
            case "show":
                showLauncher();
                break;
            case "hide":
                hideLauncher();
                break;
            case "restoreTab":
                restoreTab()
                break;
            default:
                console.warn("Unknown command:", marshalled.Command)
        }
    }
    commandSocket.onclose = () => {
        commandSocket = null;
        setTimeout(openCommandSocket, 2000)
    }

}

let rememberedTab

let showLauncher = () => {
    chrome.windows.getCurrent({}, window => {
        if (!window) {
            chrome.windows.create({ focused: true, url: "http://localhost:7938/refude/html/launcher/" })
        } else {
            chrome.tabs.query({ active: true }, ([tab]) => {
                rememberedTab = tab
                chrome.tabs.query(
                    { url: "http://localhost:7938/refude/html/launcher/" },
                    tabs => chrome.tabs.remove(tabs.map(t => t.id), () => {
                        chrome.tabs.create({ active: true, index: 0, url: "http://localhost:7938/refude/html/launcher" })
                    }))
            })
        }
    })
}

let restoreTab = () => {
    rememberedTab && chrome.tabs.update(rememberedTab.id, { active: true })

}

let hideLauncher = () => {
    chrome.tabs.query(
        { url: "http://localhost:7938/refude/html/launcher/" }, tabs => chrome.tabs.remove(tabs.map(t => t.id))
    )
}


/*
    Some nonsense one has to do to keep the service worker alive when on manifest version 3
    Stupid. And there seems to be no way of keeping alive if server is down (reconnect attempts does not 
        extend lifetime)
    Sticking to manifest v2 as long as possible

const keepAlive = () => {
    let ps = new WebSocket("ws://localhost:7938/ping")
    let ping = () => {
        if (ps) {
            console.log('pinging...')
            ps.send("ping")
            setTimeout(ping, 10000)
        }
    }
    ps.onopen = () => {
        console.log("start pinging")
        ping()
    }
    ps.onclose = () => {
        ps.close()
        ps = null
        setTimeout(keepAlive, 5000)
    }
}
keepAlive()
*/

chrome.tabs.onRemoved.addListener(reportTabs)
chrome.tabs.onUpdated.addListener(reportTabs)

openTabsSocket()
openCommandSocket()