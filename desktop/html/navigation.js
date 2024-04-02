let searchTerm = new URLSearchParams(window.location.search).get('search') || ''
let pageResource = new URLSearchParams(window.location.search).get('resource') || '/start'

let onKeyDown = event => {
    let { key, ctrlKey, altKey, shiftKey} = event;
    if (key === "Escape") {
        dismiss(true, true)
    } else if (key === 'Enter' && !ctrlKey) {
        activate()
    } else if (key === 'Delete') {
        doDelete(ctrlKey)
    } else if (key == 'Enter' && ctrlKey || key == "ArrowRight") {
        navigateTo(activeResource(), '')
    } else if (key.length === 1 && !ctrlKey && !altKey) {
        navigateTo(pageResource, searchTerm + key)
    } else if (key === "Backspace") {
        navigateTo(pageResource, searchTerm.slice(0, -1))
    } else if (key === "ArrowLeft" || key === 'h' && ctrlKey) {
        history.back() 
    } else if (key == "Tab" && !shiftKey || key == 'j' && ctrlKey || key == 'ArrowDown') {
        move()
    } else if (key == "Tab" && shiftKey || key == 'k' && ctrlKey || key == 'ArrowUp') {
        move(true)
    } else {
        return
    }
    event.preventDefault();
}


let activate = () => {
    console.log('activate:', activeProfile)
    fetch(activeResource(), {method: 'post'}).then(resp => resp.ok && dismiss())
}
let doDelete = keep => fetch(activeResource(), {method: 'delete'}).then(resp => resp.ok && !keep && dismiss(true, true))
let navigateTo = (resource, search) => window.location.search = `resource=${encodeURIComponent(resource)}&search=${encodeURIComponent(search)}`


let dismiss = (restoreWindow, restoreTab) => {  
    restoreWindow = undefined === restoreWindow ? ['window', 'browsertab'].indexOf(activeProfile()) < 0 : restoreWindow
    restoreTab = undefined === restoreTab ? 'browsertab' !== activeProfile() : restoreTab
    let searchParams = [restoreWindow && "restore=window", restoreTab && "restore=tab"].filter(e => e).join('&')
    fetch("http://localhost:7938/desktop/hide?" + searchParams, {method: 'post'})
}

let activeResource = () => document.getElementsByClassName('active')[0]?.dataset?.resource
let activeProfile = () => document.getElementsByClassName('active')[0]?.dataset?.profile

let move = up => {
    console.log("move", up)
    let trs = Array.from(document.getElementsByClassName('searchResult'))
    let len = trs.length
    if (len > 0) {
        let activeTr = document.getElementsByClassName('active')[0]
        let pos = trs.findIndex(e => e === activeTr) 
        let newPos = (pos + len + (up ? -1 : 1)) % len
        console.log("len", len, "activeTr", activeTr, "pos", pos, "newPos", newPos)
        trs.forEach((tr, i) => i === newPos ? tr.classList.add('active') : tr.classList.remove('active'))
    }
}

