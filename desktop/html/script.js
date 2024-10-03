let resourceHref = "/start"
let term = ""
let history = []

let setResourceHref = (newHref, newTerm) => {
	if (newHref !== resourceHref) {
		resourceHref = newHref
		document.getElementById("resource").dispatchEvent(new Event("resourceChanged"))
		setTerm(newTerm || "", true)
	}
}

let setTerm = newTerm => {
	console.log("set term to:", newTerm)
	term = newTerm
	document.getElementById("rows").dispatchEvent(new Event("search"))
	updateTermTag()
}

let updateTermTag = () => {
	if (document.getElementById("term")) {
		document.getElementById("term").textContent = term
	}
}

let focusFirst = () => {
	document.getElementsByClassName("title")[0]?.focus()
}

let selectedHref = () => document.activeElement?.dataset.href

let doEscape = shiftKey => {
	if (shiftKey) {
		dismiss()
	} else if (currentMenu) {
		setMenu()
	} else {
		let state = history.pop()
		if (state) {
			setResourceHref(state.path, state.term)
		} else {
			dismiss()
		}
	}
}

let goto = () => {
	if (selectedHref()) {
		history.push({ path: resourceHref, term: term })
		setResourceHref(selectedHref())
	}
}

let doEnter = ctrl => fetch(selectedHref(), { method: "post" }).then(resp => resp.ok && !ctrl && dismiss())
let doDelete = ctrl => fetch(selectedHref(), { method: "delete" }).then(resp => resp.ok && !ctrl && dismiss())
let dismiss = () => window.close()

let onKeyDown = event => {

	let { key, ctrlKey, altKey, shiftKey } = event;
	console.log("onKeyDown:", key, ctrlKey, altKey, shiftKey)

	if (key !== "Escape") {
		setMenu()
	}
	if ((key === "Escape" && !ctrlKey && !altKey) || key === 'ArrowLeft' || (key === 'o' && ctrlKey)) {
		doEscape(shiftKey)
	} else if ((key === "Enter" && altKey && !shiftKey) || key === 'ArrowRight' || (key === ' ' && ctrlKey)) {
		goto()
	} else if (key === "Enter" && !altKey && !shiftKey) {
		doEnter(ctrlKey)
	} else if (key === "Delete" && !altKey && !shiftKey) {
		doDelete(ctrlKey)
	} else if (key === "Backspace" && !ctrlKey && !altKey && !shiftKey) {
		setTerm(term.slice(0, -1))
	} else if (key.length === 1 && !ctrlKey && !altKey) {
		setTerm(term + key)
	} else if (key === "ArrowDown") {
		nextLink()?.focus()
	} else if (key === "ArrowUp") {
		nextLink('up')?.focus()
	} else {
		return
	}

	event.preventDefault();
}

let nextLink = up => {
	let current = document.activeElement?.tabIndex
	if (-1 < current) {
		newIndex = current + (up ? -1 : 1)
		return document.querySelector(`[tabindex="${newIndex}"]`)
	}
}

let doLeftClick = ev => {
	ev.stopPropagation()
	let e = ev.currentTarget
	if ("LI" === e.tagName) {
		if (e.classList.contains('submenu')) {
			toggleClosed(e)
		} else {
			execute(e.dataset.id)
		}
		return
	} else if ("IMG" === e.tagName) {
		let itemPath = e.dataset.item
		fetch(itemPath, { method: "post" }).then(resp => resp.ok && dismiss())
	}
	setMenu()
}

let doRightClick = ev => {
	ev.stopPropagation()
	ev.preventDefault()
	setMenu(ev.currentTarget.dataset.menu)
}

let currentMenu
let setMenu = menu => {
	if (menu) {
		currentMenu = menu
		document.getElementById('menu').dispatchEvent(new Event('fetchMenu'))
	} else {
		currentMenu = undefined
		document.getElementById('menu').innerHTML = ''
	}
}

let toggleClosed = element => element.classList.toggle("closed")
let execute = id => fetch(currentMenu + "?id=" + id, { method: "post" }).then(resp => { resp.ok && setMenu() })


window.addEventListener('htmx:noSSESourceError', (e) => console.log(e));
document.addEventListener("keydown", onKeyDown)
window.addEventListener("click", doLeftClick)


