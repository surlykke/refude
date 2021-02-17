// Copyright (c) Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//

import Axios from "axios";

export let getUrl = (path, handler, errHandler) => {
    Axios.get(`http://localhost:7938${path}`)
        .then(resp => handler(resp))
        .catch(err => errHandler && errHandler(err))
}

export let postUrl = (path, handler) => {
    console.log("posting:", path)
    Axios.post(`http://localhost:7938${path}`).then(resp => {
        handler && handler(resp)
    }).catch(err => console.error(err))
}

export let patchUrl = (path, body, handler) => {
    Axios.patch(`http://localhost:7938${path}`, body).then(resp => {
        handler && handler(resp)
    }).catch(err => console.error(err))
}

export let deleteUrl = (path, handler) => {
    Axios.delete(`http://localhost:7938${path}`).then(resp => {
        handler && handler(resp)
    }).catch(err => console.error(err))
}

export let addParam = (path, name, value) => {
    let separator = path.indexOf('?') > -1 ? '&' : '?'
    return path + separator + name + '=' + value
}

export let path2Url = path => "http://localhost:7938" + path

export let findLink = (res, rel, profile) => res && res._links && res._links.find(l => (l.rel === rel && (!profile || l.profile === profile)))

export let splitLinks = res => {
    let self, links
    if (res._links) {
        links = []
        res._links.forEach(l => l.rel === "self" ? self = l : links.push(l))
    }
    return {self: self, links: links}
}


export let iconClassName = link => {
    let isWindow = "/profile/window" === link.profile
    let isMinimized = isWindow && link.hints && link.hints.states && link.hints.states.indexOf("HIDDEN") > -1
    return `icon${isWindow ? " window" : ""}${isMinimized ? " minimized" : ""}`
}
