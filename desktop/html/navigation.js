
let onKeyDown = event => {
    let { key, ctrlKey, altKey, shiftKey} = event;

    if (key === "Escape") {
        dismiss()
    } else if (key === "Enter") {
        activateSelected()
    } else if (altKey && key === "l" || key === "ArrowRight") {
        navigateToSelected()
    } else if (altKey && key === "h" || key === "ArrowLeft") {
        history.back()
     } else if (key.length === 1 && !ctrlKey && !altKey || key === "Backspace")  {
        search(key)
     } else if  (altKey && key === "j" || key === "Tab" && !shiftKey || key === "ArrowDown") {
        move()
     } else if    (altKey && key === "k" || key === "Tab" && shiftKey || key === "ArrowUp") {
        move(true)
     } else {
        return
     }

    event.preventDefault();
}


let activateSelected = () => {
    let tr = selected()
    let method = tr?.dataset?.relation === "org.refude.delete" ? "delete" : "post" 
    console.log(method, "on", tr?.dataset?.href, "with profile", tr?.dataset?.profile)
    fetch(tr?.dataset?.href, {method: method}).then(resp => resp.ok && dismiss(tr?.dataset.profile))
}

let navigateToSelected = () => {
    let tr = selected()
    if (tr?.dataset?.relation === "self") {
        window.location.search = "resource=" + encodeURIComponent(tr.dataset.href)
    }
}

let search = key => {
    let searchTerm = new URLSearchParams(window.location.search).get('search') || ''
    let url = new URL(window.location)

    url.searchParams.set("search", key === "Backspace" ? searchTerm.slice(0, -1) : searchTerm + key)
    url.searchParams.delete("selected")
    window.location.replace(url)
}

let move = up => {
    let rows = Array.from(document.getElementsByClassName("selectable"))
    let pos  = rows.indexOf(document.getElementsByClassName("selected")[0]) 
    let nRows = rows.length
    if (nRows > 0) {
        let newPos = (pos + nRows + (up ? -1 : 1)) % nRows
        let url = new URL(window.location)
        url.searchParams.set("selected", newPos)
        window.location.replace(url)
    }
}

let reload = (param, newValue) => {
    let url = new URL(window.location)
    url.searchParams.set(param, newValue)
    window.location.replace(url)
}


let dismiss = ap => {  
    let restore = ap !== 'browsertab' ? (ap !== 'window' ? "window" : "tab") : ""
    fetch("http://localhost:7938/desktop/hide?restore=" + restore, {method: 'post'})
}


let selected = () => document.getElementsByClassName('selected')[0]


