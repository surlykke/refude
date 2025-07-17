const onTabUpdate = (tabId, changeInfo, tab) => {
	console.log("onTabUpdate2, changeInfo:", changeInfo)
	if ("loading" === changeInfo.status) {
		let url = new URL(tab.url)
		if ("refudecommand.localhost" === url.host) {
			let taburl = url.searchParams.get('url')
			let action = url.searchParams.get('action')
			chrome.tabs.query({ url: taburl }).then(tabs => {
				if (tabs[0]) {
					let tId = tabs[0].id
					let wId = tabs[0].windowId
					if ("focus" === action) {
						chrome.tabs.update(tId, { active: true }).then(chrome.windows.update(wId, { focused: true }))
					}
				}
			})
			chrome.tabs.remove(tabId)
		}
	} else if ("complete") {
		reportTabs()
	}
}



const onRemoved = () => {
	reportTabs()
}

const reportTabs = () => {
	chrome.tabs.query({}, tabs => {
		let data = tabs.map(t => {
			return {
				id: "" + t.id,
				title: t.title,
				url: t.url,
				favicon: t.favIconUrl
			}
		})
		console.log("data:", data)
		fetch("http://localhost:7938/browser/tabs?browserName=google-chrome", { method: "post", body: JSON.stringify(data) }).then(
			r => console.log(r)
		)
	})
}

const reportBookmarks = () => {
	/*chrome.bookmarks.getTree(bookmarks => {
		let collectedBookmarks = []

		let walk = bookmarks => bookmarks?.forEach(bookmark => {
			bookmark.id && bookmark.title && bookmark.url && collectedBookmarks.push({ id: bookmark.id, title: bookmark.title, url: bookmark.url })
			walk(bookmark.children)
		})

		walk(bookmarks)

		let msg = {
			//			browserName: browserName,
			msgType: "bookmarks",
			data: collectedBookmarks
		}
		// FIXME 
		msg.data = msg.data.length
		console.log("sending bookmarks:", msg)
	})*/
}

chrome.tabs.onRemoved.addListener(onRemoved)
chrome.tabs.onUpdated.addListener(onTabUpdate)
reportTabs()

/*if (showBookmarks) {
	chrome.bookmarks.onChanged.addListener(reportBookmarks)
	chrome.bookmarks.onCreated.addListener(reportBookmarks)
	chrome.bookmarks.onRemoved.addListener(reportBookmarks)
}*/

