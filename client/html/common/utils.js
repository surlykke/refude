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


export const retrieveResource = (path, handler, errorHandler) => {
    fetch("http://localhost:7938" + path) 
        .then(resp => resp.json())
        .then(o => handler(o))
        .catch(error => errorHandler && errorHandler(error))
}


export const followResource = (path, handler, errorHandler) => {
	follow(path, () => retrieveResource(path, handler, errorHandler), errorHandler)
}

export const retrieveCollection = (path, handler, errorHandler) => {
    fetch("http://localhost:7938" + path)
        .then(resp => resp.json())
        .then(json => Promise.allSettled(json.map(link => fetch(link.href))))
        .then(results => Promise.allSettled(results.filter(result => result.status === "fulfilled").map(result => result.value.json())))
        .then(results => handler(results.filter(r => r.status == "fulfilled").map(r => r.value)))
        .catch(error => errorHandler && errorHandler(error))
}

export const followCollection = (path, handler, errorHandler) => {
    let retrieveCollectionCurried = () => retrieveCollection(path, handler, errorHandler)
	follow(path, retrieveCollectionCurried, errorHandler)
}

export const follow = (path, handler, errorHandler) => {
   let evtSource = new EventSource("http://localhost:7938/watch")
    evtSource.onopen = () => handler()
    evtSource.onmessage = ({data}) => {
        console.log("got", data)
        data === path && handler()
    }

    evtSource.onerror = () => {
        errorHandler()
        if (evtSource.readyState === 2) {
            setTimeout(follow, 5000, path, handler, errorHandler)
        }
    } 
}

export const restorePosition = (prefix) => {
    fetch("http://localhost:7938/window/screenlayout") 
        .then(resp => resp.json())
        .then(fingerprint => {
            let screenX = parseInt(localStorage.getItem(prefix + ":" + fingerprint + '-screenX'))
            let screenY = parseInt(localStorage.getItem(prefix + ":" + fingerprint + '-screenY'))
            console.log("screenX screenY:", screenX, screenY, typeof(screenX), typeof(screenY))
            isNaN(screenX) || isNaN(screenY) || window.moveTo(Math.max(0,screenX), Math.max(0,screenY))
        })
}

export const savePositionAndClose = prefix => {
    fetch("http://localhost:7938/window/screenlayout")
        .then(resp => resp.json())
        .then(fingerprint => {
            
            localStorage.setItem(prefix + ":" + fingerprint + '-screenX', window.screenX)
            localStorage.setItem(prefix + ":" + fingerprint + '-screenY', window.screenY)
 
        })
        .finally(window.close)
}
