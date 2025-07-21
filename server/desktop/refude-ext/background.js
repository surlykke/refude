// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { browserId } from "./browserId.js"
const browserIdEncoded = encodeURIComponent(browserId)

const onTabUpdate = (tabId, changeInfo, tab) => {
	if ("loading" === changeInfo.status) {
		let url = new URL(tab.url)
		if ("refude.focustab.localhost" === url.host) {
			let taburl = url.searchParams.get('url')
			chrome.tabs.remove(tabId)
			chrome.tabs.query({ url: taburl }).then(tabs => {
				tabs.forEach((t, i) => {
					if (i === 0) {
						chrome.tabs.update(t.id, { active: true })
						chrome.windows.update(t.windowId, { focused: true })
					} else {
						chrome.tabs.remove(t.id)
					}
				})
			})
		}
	} else if ("complete" === changeInfo.staus) {
		reportTabs()
	}
}

const onTabRemove = () => {
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
		fetch(`http://localhost:7938/browser/tabs?browserId=${browserIdEncoded}`, { method: "post", body: JSON.stringify(data) })
	})
}

/*const reportBookmarks = () => {
	chrome.bookmarks.getTree(bookmarks => {
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
	})
}*/

chrome.tabs.onUpdated.addListener(onTabUpdate)
chrome.tabs.onRemoved.addListener(onTabRemove)
reportTabs()
/*
chrome.bookmarks.onChanged.addListener(reportBookmarks)
chrome.bookmarks.onCreated.addListener(reportBookmarks)
chrome.bookmarks.onRemoved.addListener(reportBookmarks)
reportBoookmarks()
*/

