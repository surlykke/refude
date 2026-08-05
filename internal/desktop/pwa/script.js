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
	handle(e.key.length === 1 && !e.ctrlKey, () => setTerm(term + e.key)) ||
	handle(e.key === "Enter" && e.ctrlKey, () =>  { e.target.dispatchEvent(new Event("toggledetails", {bubbles: true}))}) || 
	handle(e.key === "Enter" && !e.ctrlKey, () => doPost(e.target)) ||
	handle(e.key === "Escape", () => term ? setTerm("") : window.close()) ||
	handle(e.key === "Backspace", () => setTerm(term.slice(0, -1)))
}

let doPost = el => el.dataset.post && fetch(el.dataset.post, {method: "POST"}).then(resp => resp.ok && window.close())
let doDelete = el => el.dataset.delete && fetch(el.dataset.delete, {method: "DELETE"}).then(resp => resp.ok && window.close())
let afterSettle = () => document.getElementsByClassName('title')[0]?.focus()

