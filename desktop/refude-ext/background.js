let tabsSocket
let commandSocket

const reportTabs = () => {
	chrome.tabs.query({}, tabs => {
		let tabsData = tabs.map(t => {
			return {
				id: "" + t.id,
				title: t.title,
				url: t.url,
				favIcon: t.favIconUrl
			}
		})
		fetch("http://localhost:7938/tabsink?browserName=" + browserName, { method: "POST", body: JSON.stringify(tabsData) })
			.then(response => {
				if (!response.ok) {
					throw new Error(response.status)
				}
			})
			.catch(() => { // If we couldn't deliver data, try again i 5 secs.
				setTimeout(reportTabs, 5000)
			})
	})
}

const reportBookmarks = () => {
	
	chrome.bookmarks.getTree(bookmarks => {
		let collectedBookmarks = []

		let walk = bookmarks => bookmarks?.forEach(bookmark => {
			bookmark.id && bookmark.title && bookmark.url && collectedBookmarks.push({id: bookmark.id, title: bookmark.title, url: bookmark.url})	
			walk(bookmark.children)
		})

		walk(bookmarks, collectedBookmarks)

		fetch("http://localhost:7938/bookmarksink", { method: "POST", body: JSON.stringify(collectedBookmarks) })
			.then(response => {
				if (!response.ok) {
					throw new Error(response.status)
				}
			})
			.catch(() => { // If we couldn't deliver data, try again i 5 secs.
				setTimeout(reportBookmarks, 5000)
			})
	})
}


const watch = () => {
	let evtSource = new EventSource("http://localhost:7938/watch")
	evtSource.onopen = () => {
		reportTabs()
		reportBookmarks()
	}
	evtSource.addEventListener("focusTab", ({ data }) => {
		focusTab(parseInt(data))
	})
	evtSource.addEventListener("closeTab", ({ data }) => {
		let tabId = parseInt(data)
		tabId && chrome.tabs.remove(tabId)
	})
	evtSource.onerror = error => {
		console.log(error)
		if (evtSource.readyState === 2) {
			setTimeout(watch, 5000)
		}
	}
}


let focusTab = tabId => {
	chrome.tabs.get(tabId, tab => {
		chrome.tabs.update(tab.id, { active: true })
		chrome.windows.update(tab.windowId, { focused: true })
	})
}

console.log("browser name:", browserName)

reportTabs()
chrome.tabs.onRemoved.addListener(reportTabs)
chrome.tabs.onUpdated.addListener(reportTabs)

if (showBookmarks) {
	reportBookmarks()
	chrome.bookmarks.onChanged.addListener(reportBookmarks)
	chrome.bookmarks.onCreated.addListener(reportBookmarks)
	chrome.bookmarks.onRemoved.addListener(reportBookmarks)
}

watch()
