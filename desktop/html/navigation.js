let searchTerm = new URLSearchParams(window.location.search).get('search') || ''
let pageResource = new URLSearchParams(window.location.search).get('resource') || '/start'

let onKeyDown = event => {
    let { key, ctrlKey, altKey, shiftKey} = event;
    console.log(key)
    if (key === "Escape") {
        dismiss()
    } else if (key === 'Enter' && !ctrlKey) {
        activate(selected())
    } else if (key === 'Enter' && ctrlKey || key === "ArrowRight" || key === "i" && ctrlKey) {
        navigateTo(selected())
    } else if (key === "ArrowLeft" || key === 'h' && ctrlKey || key === "b" && ctrlKey) {
        history.back()
    } else if (key.length === 1 && !ctrlKey && !altKey) {
        search(searchTerm + key)
    } else if (key === "Backspace") {
        search(searchTerm.slice(0, -1))
    } else if (key === "Tab" && !shiftKey || key === 'j' && ctrlKey || key === 'ArrowDown') {
        move()
    } else if (key === "Tab" && shiftKey || key === 'k' && ctrlKey || key === 'ArrowUp') {
        move(true)
    } else {
        return
    }
    event.preventDefault();
}


let activate = tr => {
    if (!tr || !method(tr)) return 
    console.log("activate, tr", tr, ",method", method, "href", tr.dataset?.href)
    fetch(tr.dataset.href, {method: method(tr)}) .then(resp => resp.ok && dismiss(tr.dataset.profile))
}

let navigateTo = tr => {
    if (!tr || !tr.dataset.href || !canGet(tr)) return
    window.location.search = "resource=" + encodeURIComponent(tr.dataset.href)
}
    
let search = searchTerm => {
    let url = new URL(window.location)
    url.searchParams.set("search", searchTerm)
    window.location.replace(url)
}

let dismiss = ap => {  
    let restore = ap !== 'browsertab' ? (ap !== 'window' ? "window" : "tab") : ""
    fetch("http://localhost:7938/desktop/hide?restore=" + restore, {method: 'post'})
}

let move = up => {
    let trs = selectables()
    let len = trs.length
    let pos = trs.indexOf(selected())
    pos = pos < 0 ? 0 : (pos + len + (up ? -1 : 1)) % len
    trs.forEach((tr, i) => i === pos ? tr.classList.add('selected') : tr.classList.remove('selected'))
}

let selectables = () => Array.from(document.getElementsByClassName('selectable'))
let selected = () => document.getElementsByClassName('selected')[0]
let method = tr => {
    let rel = tr.dataset.relation || ""
    let prof = tr.dataset.profile
    return rel === "org.refude.delete" ? "delete" : 
           rel === "org.refude.action" || ["window", "browsertab", "application", "notification"].includes(prof) ? "post" :
           "none"
}

let canDelete = tr => ["window" /*TODO*/].includes(tr.dataset.profile)
let canGet = tr => tr.dataset.relation === "self"

selectables()[0]?.classList.add('selected')
