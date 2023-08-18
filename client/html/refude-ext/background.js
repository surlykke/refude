let websocket
let notifications = []

let id = notification => "refude-notification-" + notification.NotificationId

const handler = m => {
    let newNotifications = JSON.parse(m.data)
    for (let n of notifications) {
        if (!newNotifications.find(nn => nn.NotificationId === n.NotificationId)) {
            chrome.notifications.clear(id(n))
        }
    }

    for (let nn of newNotifications) {
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
    websocket.close()
    setTimeout(consumeNotifications, 5000)
}

const consumeNotifications = () => {
    console.log("Connecting")
    websocket = new WebSocket("ws://localhost:7938/notification/websocket")
    websocket.addEventListener('message', handler)
    websocket.addEventListener('error', () => console.warn("Error:", error))
    websocket.addEventListener('close', () => {
        websocket.close()
        setTimeout(consumeNotifications, 5000)
    })
}

// launch_handler "focus-existing" doesn't work for tabbed pwa's, so this instead
// But a bit more agressive.. Close existing launcher tabs whenever a new tab is opened, so
// launcher tabs becomes ephemeral (kind of..)
const handleTabCreated = tab => {
    chrome.tabs.query(
        { url: "http://localhost:7938/refude/html/launcher/" },
        tabs => chrome.tabs.remove(tabs.map(t => t.id).filter(id => id !== tab.id))
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

chrome.tabs.onCreated.addListener(handleTabCreated)
consumeNotifications("/notification/", handler, error => console.log(error))
