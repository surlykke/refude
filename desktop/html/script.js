let term = ""

let setTerm = newTerm => {
	document.getElementById("term").textContent = term = newTerm
	search()
}

let search = () => document.getElementById("search-results").dispatchEvent(new Event("search"))

let selectedPath = () => document.activeElement?.dataset.path

let doEscape = shiftKey => {
	if (shiftKey) {
		dismiss()
	} else if (term) {
		setTerm("")
	} else {
		dismiss()
	}
}

let doEnter = (ctrl, shift) => {
	if (ctrl) {
		return
	} else {
		href = document.activeElement?.dataset.href
		rel = document.activeElement?.dataset.rel

		console.log("href:", href, "rel:", rel)
		if (rel === "org.refude.action") {
			fetch(href, { method: "post" }).then(resp => resp.ok && !shift && dismiss())
		}
	}
}

let dismiss = () => window.close()

let onKeyDown = event => {

	let { key, ctrlKey, altKey, shiftKey } = event;
	console.log("onKeyDown:", key, ctrlKey, altKey, shiftKey)

	if ((key === "Escape" && !ctrlKey && !altKey) || (key === 'o' && ctrlKey)) {
		doEscape(shiftKey)
	} else if (key === "Enter" && !altKey) {
		doEnter(ctrlKey, shiftKey)
	} else if (key === "Delete" && !altKey && ctrlKey) {
		doDelete(shiftKey)
	} else if (key === "Backspace" && !ctrlKey && !altKey && !shiftKey) {
		setTerm(term.slice(0, -1))
	} else if (key.length === 1 && !ctrlKey && !altKey) {
		setTerm(term + key)
	} else if (key === "ArrowDown" || (ctrlKey && key === 'j')) {
		nextLink()?.focus()
	} else if (key === "ArrowUp" || (ctrlKey && key === 'k')) {
		nextLink('up')?.focus()
	} else if (key === "ArrowRight" || (ctrlKey && key === 'l')) {
		document.activeElement?.nextElementSibling?.focus()
	}  else if (key === "ArrowLeft" || (ctrlKey && key === 'h')) {
		document.activeElement?.previousElementSibling?.focus()
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

let toggleClosed = element => element.classList.toggle("closed")

window.addEventListener('htmx:noSSESourceError', (e) => console.log(e));
document.addEventListener("keydown", onKeyDown)


