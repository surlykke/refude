// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
//
let term = ""

let setTerm = t => {
	term = t
	document.getElementById("term").innerHTML = term
	document.getElementById("search-results").dispatchEvent(new Event("search"))
}

let onKeyDown = e => {
	let handle = (condition, action) => {
		if (!condition) return false
		action()
		e.preventDefault()
		return true
	}
	console.log(e)
	handle(e.key.length === 1 && !e.ctrlKey, () => setTerm(term + e.key)) ||
		handle((e.key == "j" || e.key == "n") && e.ctrlKey, () => step(e.target)) ||
		handle((e.key == "k" || e.key == "p") && e.ctrlKey, () => step(e.target, true)) ||
		handle(e.key == "Tab" && !e.shiftKey, () => step(e.target)) ||
		handle(e.key == "Tab" && e.shiftKey, () => step(e.target, true)) ||
		handle(e.key === "Enter" && e.ctrlKey, () => { e.target.dispatchEvent(new Event("getdetails", { bubbles: true })) }) ||
		handle(e.key === "Enter" && !e.ctrlKey, () => doPost(e.target)) ||
		handle(e.key === "Escape", () => term ? setTerm("") : window.close()) ||
		handle(e.key === "Backspace", () => setTerm(term.slice(0, -1)))
}

let doPost = el => el.dataset.post && fetch(el.dataset.post, { method: "POST" }).then(resp => resp.ok && window.close())
let doDelete = el => el.dataset.delete && fetch(el.dataset.delete, { method: "DELETE" }).then(resp => resp.ok && window.close())
let afterSettle = () => document.getElementsByClassName('title')[0]?.focus()

let afterActionsSettle = e => e.querySelectorAll("[tabindex]")[0]?.focus()

let step = (el, backward) => {
	let focusable = Array.from(document.querySelectorAll("[tabindex]"))
	let idx = focusable.indexOf(el)
	focusable[(idx + (backward ? -1 : 1) + focusable.length) % focusable.length]?.focus()
}

