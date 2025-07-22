// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
let term = ""

let setTerm = newTerm => {
	document.getElementById("term").textContent = term = newTerm
	document.getElementById("search-results").dispatchEvent(new Event("search"))
}

let doCtrlSpace = () => {
	document.activeElement?.dispatchEvent(new Event("details"))
}

let doEscape = shiftKey => {
	if (shiftKey) {
		dismiss()
	} else if (document.querySelector(".action")) {
		setTerm(term)
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
		if (href) {
			fetch(href, { method: "post" }).then(resp => resp.ok && !shift && dismiss())
		}
	}
}

let setTabIndexes = () => {
	document.querySelectorAll('[data-href]').forEach((e, i) => e.tabIndex = i + 1)
	document.activeElement?.hasAttribute('tabindex') || document.querySelector('[tabIndex="1"]')?.focus()
}

let dismiss = () => window.close()

let onKeyDown = event => {

	let { key, ctrlKey, altKey, shiftKey } = event;

	if ((key === "Escape" && !ctrlKey && !altKey) || (key === 'o' && ctrlKey)) {
		doEscape(shiftKey)
	} else if (key === "Enter" && !altKey) {
		doEnter(ctrlKey, shiftKey) 
	} else if (key === " " && ctrlKey) {
		doCtrlSpace()
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

document.addEventListener("keydown", onKeyDown)


