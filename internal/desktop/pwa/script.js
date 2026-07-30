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
	handle(e.key === "Enter" && e.ctrlKey, () =>  {
		console.log("Dispatch toggleShowDetails to", e.target)
		e.target.dispatchEvent(new Event("toggleshowdetails"))
	})|| 
	handle(e.key === "Enter" && !e.ctrlKey, () => e.target.dispatchEvent(new Event('activate'))) ||
	handle(e.key === "Escape", () => term ? setTerm("") : window.close()) ||
	handle(e.key === "Backspace", () => setTerm(term.slice(0, -1)))
}

let afterSettle = e => {
	if ("/desktop/search" === e.detail.pathInfo.requestPath) {
		document.getElementsByClassName('title')[0]?.focus()
	} else if ("post" === e.detail.requestConfig.verb && e.detail.successful) {
		window.close()
	}
}

let toggleShowDetails = e => {
	console.log("We are toggling")
	let div = e.target.nextElementSibling.nextElementSibling
	div.style.display = div.style.display === 'none' ? '' : 'none'
}

