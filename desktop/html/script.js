let term = ""
let details = ""

let setTerm = newTerm => {
	document.getElementById("term").textContent = term = newTerm
	search()
}

let setDetails = newDetails => {
	details = newDetails
	search()
}


let search = () => document.getElementById("search-results").dispatchEvent(new Event("search"))

let selectedPath = () => document.activeElement?.dataset.path
let doEscape = shiftKey => {
	if (shiftKey) {
		dismiss()
	} else if (currentMenu) {
		setMenu()
	} else if (details) {
		setDetails("")
	} else if (term) {
		setTerm("")
	} else {
		dismiss()
	}
}

let doEnter = (ctrl, shift) => {
	console.log("doEnter", ctrl, shift)
	if (ctrl && shift) {
		return
	} else if (ctrl) {
		let newDetails = selectedPath() && selectedPath() !== details ? selectedPath() : ""
		setDetails(newDetails)
	} else {
		fetch(selectedPath(), { method: "post" }).then(resp => resp.ok && !shift && dismiss())
	}
}

let doDelete = shift => fetch(selectedPath(), { method: "delete" }).then(resp => resp.ok && !shift && dismiss())
let dismiss = () => window.close()

let onKeyDown = event => {

	let { key, ctrlKey, altKey, shiftKey } = event;
	console.log("onKeyDown:", key, ctrlKey, altKey, shiftKey)

	if (key !== "Escape") {
		setMenu()
	}
	if ((key === "Escape" && !ctrlKey && !altKey) || key === 'ArrowLeft' || (key === 'o' && ctrlKey)) {
		doEscape(shiftKey)
	} else if (key === "Enter" && !altKey) {
		doEnter(ctrlKey, shiftKey)
	} else if (key === "Delete" && !altKey && ctrlKey) {
		doDelete(shiftKey)
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
			fetch(currentMenu + "?id=" + e.dataset.id, { method: "post" }).then(resp => resp.ok && setMenu())
		}
		return
	} else if ("IMG" === e.tagName) {
		let itemPath = e.dataset.item
		fetch(itemPath, { method: "post" })
	}
	setMenu()
}

let doRightClick = ev => {
	ev.stopPropagation()
	ev.preventDefault()
	console.log("doRightClick, currentTarget:", ev.currentTarget)
	setMenu(ev.currentTarget.dataset.menu)
}

let currentMenu
let setMenu = menu => {
	if (menu) {
		document.getElementById('menu').style.display = "block"
		currentMenu = menu
		document.getElementById('menu').dispatchEvent(new Event('fetchMenu'))
	} else {
		currentMenu = undefined
		document.getElementById('menu').innerHTML = ''
		document.getElementById('menu').style.display = "none"
	}
}

let toggleClosed = element => element.classList.toggle("closed")


window.addEventListener('htmx:noSSESourceError', (e) => console.log(e));
document.addEventListener("keydown", onKeyDown)
window.addEventListener("click", doLeftClick)


