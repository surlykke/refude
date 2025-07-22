// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
var port = chrome.runtime.connectNative('org.refude.native_messaging');

const onTabUpdate = (_tabId, changeInfo, _tab) => {
	if ("complete" === changeInfo.staus) {
		reportTabs()
	}
}

const onTabRemove = () => {
	reportTabs()
}

const reportTabs = () => {
	chrome.tabs.query({}, tabs => {
		let data = {
			type: "tabs",
			list: tabs.map(t => {
				return {
					id: "" + t.id,
					title: t.title,
					url: t.url,
					favicon: t.favIconUrl
				}
			})
		}
		port.postMessage(data)
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


port.onMessage.addListener(function(obj) {
	let tabId = parseInt(obj.tabId)
	if (tabId && "focus" === obj.cmd) {
		chrome.tabs.get(tabId).then(t => {
			chrome.tabs.update(tabId, { active: true })
			chrome.windows.update(t.windowId, { focused: true })
		})
	}
});

port.onDisconnect.addListener(function() {
	// TODO Set up alarm
	console.log('Disconnected');
});

chrome.tabs.onUpdated.addListener(onTabUpdate)
chrome.tabs.onRemoved.addListener(onTabRemove)
reportTabs()
/*
chrome.bookmarks.onChanged.addListener(reportBookmarks)
chrome.bookmarks.onCreated.addListener(reportBookmarks)
chrome.bookmarks.onRemoved.addListener(reportBookmarks)
reportBoookmarks()
*/

