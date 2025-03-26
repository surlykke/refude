let socket

const reportTabs = () => {
    chrome.tabs.query({}, tabs => {
        if (socket) {
			let msg = {
				browserName: browserName,
				msgType: "tabs",
				data:tabs.map(t => {
					return {
						id: "" + t.id,
						title: t.title,
						url: t.url,
						favicon: t.favIconUrl 
					}
				})
		 	}
            socket.send(JSON.stringify(msg))
        } 
	})
}

const reportBookmarks = () => {
	if (socket) {	
	chrome.bookmarks.getTree(bookmarks => {
		let collectedBookmarks = []

		let walk = bookmarks => bookmarks?.forEach(bookmark => {
			bookmark.id && bookmark.title && bookmark.url && collectedBookmarks.push({id: bookmark.id, title: bookmark.title, url: bookmark.url})	
			walk(bookmark.children)
		})

		walk(bookmarks, collectedBookmarks)

		let msg = {
			browserName: browserName,
			msgType: "bookmarks",
			data: collectedBookmarks
		}
		socket.send(JSON.stringify(msg))

	})	
	}
}

const openSocket = () => {
    socket = new WebSocket("ws://localhost:7938/browser/socket") // ?
    socket.onopen = () => { 
		reportTabs(); 
		if (showBookmarks) {
			reportBookmarks()
		}
	}
    socket.onmessage = m => {
        let msg = JSON.parse(m.data)
        let tabId = parseInt(msg.tabId)
        if (tabId) {
			if (msg.operation === "focus") {
				chrome.tabs.get(tabId, tab => {
					chrome.tabs.update(tabId, { 'active': true })
                	chrome.windows.update(tab.windowId, { focused: true })
				})
			} else if (operation === "close") {
				chrome.tabs.close(tabId)
			}
        }
    }
    socket.onclose = () => {
        socket = null;
        setTimeout(openSocket, 10000)
    }
}

openSocket()

chrome.tabs.onRemoved.addListener(reportTabs)
chrome.tabs.onUpdated.addListener(reportTabs)

if (showBookmarks) {
	chrome.bookmarks.onChanged.addListener(reportBookmarks)
	chrome.bookmarks.onCreated.addListener(reportBookmarks)
	chrome.bookmarks.onRemoved.addListener(reportBookmarks)
}

