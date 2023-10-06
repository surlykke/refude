let notificationSocket
let notifications = []

let tabsSocket

let id = notification => "refude-notification-" + notification.NotificationId

const handler = m => {
    let newNotifications = JSON.parse(m.data)
    for (let n of notifications) {
        if (!newNotifications.find(nn => nn.NotificationId === n.NotificationId)) {
            chrome.notifications.clear(id(n))
        }
    }

    for (let nn of newNotifications) {
        if (!nn.IconName) {
            nn.IconName = nn.Urgency === 2 ? "dialog-warning" : "dialog-info"
        }
        chrome.notifications.create(
            id(nn), {
            type: "basic",
            iconUrl: `http://localhost:7938/icon?name=${nn.IconName}`,
            title: nn.Title || " ",
            message: nn.Comment || "",
            priority: nn.Urgency,
            requireInteraction: true
        })
    }
    notifications = newNotifications
}

const errorHandler = error => {
    console.log("Error:", error)
}

const closeHandler = () => {
    console.log("Connection closed")
    notificationSocket.close()
    setTimeout(consumeNotifications, 5000)
}

const consumeNotifications = () => {
    notificationSocket = new WebSocket("ws://localhost:7938/notification/websocket")
    notificationSocket.addEventListener('message', handler)
    notificationSocket.addEventListener('error', error => console.log("Error:", error))
    notificationSocket.addEventListener('close', () => {
        notificationSocket.close()
        setTimeout(consumeNotifications, 5000)
    })
}

// launch_handler "focus-existing" doesn't work for tabbed pwa's, so this instead
// But a bit more agressive.. Close existing launcher tabs whenever a new tab is opened, so
// launcher tabs becomes ephemeral (kind of..)
const cleanUpLauncherTabs = tab => {
    chrome.tabs.query(
        { url: "http://localhost:7938/refude/html/launcher/" },
        tabs => {
            chrome.tabs.remove(tabs.map(t => t.id).filter(id => id !== tab.id))
        }
    )
}

const reportTabs = () => {
    console.log("Reporting tabs")
    chrome.tabs.query({}, tabs => {
        console.log("tabsSocket:", tabsSocket)
        console.log("tabids:", tabs.map(t => t.id).join(", "))
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
        console.log("m.data:", m.data)
        let marshalled = JSON.parse(m.data)
        let tabId = parseInt(marshalled)
        console.log("tabId:", tabId)
        if (tabId) {
            var updateProperties = { 'active': true };
            chrome.tabs.update(tabId, updateProperties, (tab) => { }) 
        }
    }
    tabsSocket.onclose = () => {
        tabsSocket = null;
        setTimeout(openTabsSocket, 10000)
    }
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

chrome.tabs.onCreated.addListener(cleanUpLauncherTabs)
chrome.tabs.onRemoved.addListener(reportTabs)
chrome.tabs.onUpdated.addListener(reportTabs)

consumeNotifications("/notification/", handler, error => console.log(error))
openTabsSocket()