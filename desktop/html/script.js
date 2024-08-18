let resourcePath = "/start"
let term = ""
let history = []

let setResourcePath = (newPath, newTerm) => {
	if (newPath !== resourcePath) {
		resourcePath = newPath
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

let selectedDataset = () => document.activeElement.dataset || {}

let doEscape = shiftKey => {
	let state = history.pop()
	if (state && !shiftKey) {
		setResourcePath(state.path, state.term)
	} else {
		dismiss()
	}
}

let goto = () => {
	console.log("goto, dataset:", selectedDataset())
	let { get } = selectedDataset()
	if (get) {
		history.push({ path: resourcePath, term: term })
		setResourcePath(get)
	}
}

let doEnter = ctrl => {
	if (selectedDataset().post) {
		fetch(selectedDataset().post, { method: "post" })
			.then(resp => resp.ok && !ctrl && dismiss(selectedDataset().profile))
	}
}

let doDelete = ctrl => {
	if (selectedDataset().delete) {
		fetch(selectedDataset().delete, { method: "delete" })
			.then(resp => resp.ok && !ctrl && dismiss(selectedDataset().profile))
	}
}


let dismiss = actionProfile => {
	console.log("Dismiss")
	let restore = actionProfile !== 'tab' ? (actionProfile !== 'window' ? "window" : "tab") : ""
	console.log("Doing post against http://localhost:7938/desktop/hide?restore=" + restore)
	fetch("http://localhost:7938/desktop/hide?restore=" + restore, { method: 'post' })
}



let onKeyDown = event => {
	let { key, ctrlKey, altKey, shiftKey } = event;
	if (key === "Escape" && !ctrlKey && !altKey) {
		doEscape(shiftKey)
	} else if (key === "Enter" && altKey && !shiftKey) {
		goto()
	} else if (key === "Enter" && !altKey && !shiftKey) {
		doEnter(ctrlKey)
	} else if (key === "Delete" && !altKey && !shiftKey) {
		doDelete(ctrlKey)
	} else if (key == "Backspace" && !ctrlKey && !altKey && !shiftKey) {
		setTerm(term.slice(0, -1))
	} else if (key.length === 1 && !ctrlKey && !altKey) {
		setTerm(term + key)
	} else {
		return
	}

	event.preventDefault();
}


window.addEventListener('htmx:noSSESourceError', (e) => {
	console.log(e);
});

document.addEventListener("keydown", onKeyDown)
//document.addEventListener("readystatechange", () => document.readyState === "complete" && searchTag().dispatchEvent(new Event("termChanged")))
