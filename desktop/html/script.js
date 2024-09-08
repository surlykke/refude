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
	document.getElementsByClassName("row")[0]?.getElementsByTagName('a')[0]?.focus()
}

let selectedHref = () => document.activeElement?.href

let doEscape = shiftKey => {
	let state = history.pop()
	if (state && !shiftKey) {
		setResourceHref(state.path, state.term)
	} else {
		dismiss()
	}
}

let goto = () => {
	console.log("goto, selectedHref:", selectedHref())
	if (selectedHref()) {
		history.push({ path: resourceHref, term: term })
		setResourceHref(selectedHref())
	}
}

let doEnter = ctrl => {
	console.log("selectedHref:", selectedHref())

	fetch(selectedHref(), { method: "post" }).then(resp => resp.ok && !ctrl && dismiss())
}

let doDelete = ctrl => fetch(selectedHref(), { method: "delete" }).then(resp => resp.ok && !ctrl && dismiss())

let dismiss = () => window.close()

let onKeyDown = event => {
	let { key, ctrlKey, altKey, shiftKey } = event;
	if ((key === "Escape" && !ctrlKey && !altKey) || key === 'ArrowLeft') {
		doEscape(shiftKey)
	} else if ((key === "Enter" && altKey && !shiftKey) || key === 'ArrowRight') {
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
		nextLink().focus()
	} else if (key === "ArrowUp") {
		nextLink('up').focus()
	} else {
		return
	}

	event.preventDefault();
}

let nextLink = up => {
	let a = Array.from(document.getElementsByClassName('link'))
	return a[a.indexOf(document.activeElement) + (up ? -1 : 1)];
}


window.addEventListener('htmx:noSSESourceError', (e) => console.log(e));

document.addEventListener("keydown", onKeyDown)
