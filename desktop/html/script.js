const selectables = document.getElementsByClassName('selectable')
const selected = document.getElementsByClassName('selected')

let state = {res: "/start", term: "",  pos: 0}
let history = []
let updated 


let load = () => {
    let url = `/desktop/body?resource=${state.res}&search=${state.term}`
    fetch(url)
        .then(r => r.ok ? r.text() : Promise.reject())
        .then(text => {
            document.body.innerHTML = text
            highlightSelected()
            updated = parseInt(document.getElementById('table')?.dataset?.updated)
        })
}

let highlightSelected = () => Array.from(selectables).forEach((e, i) => i === state.pos ? e.classList.add('selected') : e.classList.remove('selected'))


let gotoResource = newResource => {
    console.log("gotoResource:", newResource)
    if (newResource) {
        history.push(state)
        state = {res: newResource, term: '', pos: 0}
        load()
    }
}

let goBack = () => {
    state = history.pop() || {res: '/start', term: "", pos: 0}
    load()
}

let setTerm = newTerm => {
    state.term = newTerm
    state.pos = 0
    load()
}


let selectedDataset = () => selected.item(0)?.dataset

let activateSelected = () => {
    if (!selectedDataset()) return 
    let method = selectedDataset().relation === "org.refude.delete" ? "delete" : "post"
    let profile = selectedDataset().profile
    fetch(selectedDataset().href, { method: method }).then(resp => resp.ok && dismiss(profile))
}

let onKeyDown = event => {
    let { key, ctrlKey, altKey, shiftKey } = event;

    if (key === "Escape") {
        dismiss()
    } else if (key === "Enter") {
        activateSelected()
    } else if (altKey && key === "l" || key === "ArrowRight") {
        console.log("rel:", selectedDataset()?.relation)
        selectedDataset()?.relation === "self" && gotoResource(selectedDataset().href)
    } else if (altKey && key === "h" || key === "ArrowLeft") {
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
    let restore = actionProfile !== 'browsertab' ? (actionProfile !== 'window' ? "window" : "tab") : ""
    fetch("http://localhost:7938/desktop/hide?restore=" + restore, { method: 'post' })
}

document.addEventListener("keydown", onKeyDown)
load()

let watch = () => {
    if (document.visibilityState === 'hidden') return
    fetch("/desktop/lastupdate")
        .then(r => r.ok && r.json())
        .then(upd => upd > updated && load())
}
setInterval(watch, 300)