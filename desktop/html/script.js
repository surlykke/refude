const selectables = document.getElementsByClassName('selectable')
const selected = document.getElementsByClassName('selected')
let etag = '""'

let state = { res: "/start", term: "", pos: 0 }
let pageHistory = []
let hash = ""


let load = () => {
	let url = `/desktop/body?resource=${state.res}&search=${state.term}`
	fetch(url, { headers: { "If-None-Match": etag } })
		.then(response => {
			if (response.ok) {
				etag = response.headers.get("ETag");
				return response.text()
			} else {
				return Promise.reject()
			}
		})
		.then(text => {
			document.body.innerHTML = text
			highlightSelected()
			hash = document.getElementById('table')?.dataset?.hash
		})
		.catch(e => { })
}

let highlightSelected = () => {
	Array.from(selectables).forEach(e => e.classList.remove('selected'))
	let item = selectables.item(state.pos)
	if (item) {
		item.classList.add('selected')
		let list = document.getElementsByClassName('rows')[0]
		let listRect = list.getBoundingClientRect()
		let itemRect = item.getBoundingClientRect()
		let neededScroll = itemRect.y + itemRect.height + list.scrollTop - (listRect.y + listRect.height)
		neededScroll = neededScroll < 0 ? 0 : neededScroll
		list.scroll(0, neededScroll)
	}
}

let gotoResource = newResource => {
	if (newResource) {
		pageHistory.push(state)
		state = { res: newResource, term: '', pos: 0 }
		load()
	}
}

let goBack = () => {
	state = pageHistory.pop() || { res: '/start', term: "", pos: 0 }
	load()
}

let setTerm = newTerm => {
	state.term = newTerm
	state.pos = 0
	load()
}


let selectedDataset = () => selected.item(0)?.dataset

let activateSelected = () => {
	let dataset = selected.item(0)?.dataset
	if (!dataset) return
	let profile = dataset.profile
	switch (dataset.relation) {
		case "self":
			console.log("Fetch", dataset.href)
			fetch(dataset.href)
				.then(resp => resp.json())
				.then(jsonMap => {
					let defaultAction = jsonMap?.links?.find(l => l.rel === "org.refude.action")
					if (defaultAction) {
						fetch(defaultAction.href, { method: "post" }).then(resp => resp.ok && dismiss(profile))
						console.log(defaultAction)
					}
				})
			break
		case "org.refude.action":
			fetch(dataset.href, { method: "post" }).then(resp => resp.ok && dismiss(profile))
			break
		case "org.refude.delete":
			fetch(dataset.href, { method: "delete" }).then(resp => resp.ok && dismiss(profile))
			break
	}
}

let onKeyDown = event => {
	let { key, ctrlKey, altKey, shiftKey } = event;
	if (key === "Escape") {
		dismiss()
	} else if (key === "Enter" && !ctrlKey && !shiftKey && !altKey) {
		activateSelected()
	} else if (altKey && key === "l" || key === "ArrowRight" || key === " " && ctrlKey) {
		selectedDataset()?.relation === "self" && gotoResource(selectedDataset().href)
	} else if (altKey && key === "h" || key === "ArrowLeft" || key === "o" && ctrlKey) {
		goBack()
	} else if (key.length === 1 && !ctrlKey && !altKey) {
		setTerm(state.term + key)
	} else if (key === "Backspace") {
		setTerm(state.term.slice(0, -1))
	} else if (altKey && key === "j" || key === "Tab" && !shiftKey || key === "ArrowDown") {
		move()
	} else if (altKey && key === "k" || key === "Tab" && shiftKey || key === "ArrowUp") {
		move(true)
	} else {
		return
	}

	event.preventDefault();
}


let move = up => {
	state.pos = selectables.length === 0 ? 0 : (state.pos + selectables.length + (up ? -1 : 1)) % selectables.length
	highlightSelected()
}

let dismiss = actionProfile => {
	window.location.search = ''
	let restore = actionProfile !== 'tab' ? (actionProfile !== 'window' ? "window" : "tab") : ""
	fetch("http://localhost:7938/desktop/hide?restore=" + restore, { method: 'post' })
}

document.addEventListener("keydown", onKeyDown)
load()

let reloadOnChange = () => {
	if (document.visibilityState === 'visible') {
		load()
	}
}
//setInterval(reloadOnChange, 500)
