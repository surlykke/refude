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
		fetch("http://localhost:7938/tabsink", { method: "POST", body: JSON.stringify(tabsData) })
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

const watch = () => {
	let evtSource = new EventSource("http://localhost:7938/watch")
	evtSource.onopen = reportTabs
	evtSource.addEventListener("focusTab", ({ data }) => {
		focusTab(parseInt(data))
	})
	evtSource.addEventListener("closeTab", ({ data }) => {
		console.log("closeTab", data)
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


reportTabs()
chrome.tabs.onRemoved.addListener(reportTabs)
chrome.tabs.onUpdated.addListener(reportTabs)
watch()
