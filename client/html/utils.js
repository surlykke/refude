// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
export const iconClassName = profile => "icon" + ("window" === profile ? " window" : "")

export const doPost = (href, params) => {
    if (params) {
        let separator = href.indexOf('?') > -1 ? '&' : '?'
        for (const [name, val] of Object.entries(params)) {
            href = href + separator + name + "=" + val
            separator = "&"
        }
    }
    return fetch(href, { method: "POST" })
}

export const doDelete = href => fetch(href, { method: "DELETE" })

export const watchResource = (path, handler) => {
    let evtSource = new EventSource("http://localhost:7938/watch")
    evtSource.onmessage = ({data}) => data === path && handler()
    evtSource.onerror = () => {
        errorHandler()
        if (evtSource.readyState === 2) {
            setTimeout(followResource, 5000, path, handler, errorHandler)
        }
    }
}

export const followResource = (path, handler, errorHandler) => {
    let retrieveResource = () => {
        console.log("Retrieving", path)
        fetch("http://localhost:7938" + path) 
            .then(resp => resp.json())
            .then(o => {console.log("handing:", o, "to handler"); handler(o)}, error => errorHandler && errorHandler(error))
    }

   let evtSource = new EventSource("http://localhost:7938/watch")
    evtSource.onopen = () => retrieveResource()
    evtSource.onmessage = ({data}) => data === path && retrieveResource()

    evtSource.onerror = () => {
        errorHandler()
        if (evtSource.readyState === 2) {
            setTimeout(followResource, 5000, path, handler, errorHandler)
        }
    } 

}
