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
    console.log("Axios.post", `http://localhost:7938${path}`)
    Axios.post(`http://localhost:7938${path}`).then(resp => {
        console.log("resp...")
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

export let iconClassName = link => {
    let isWindow = link.traits && link.traits.includes("window")
    let isMinimized = isWindow && link.traits.includes("minimized")
    return `icon${isWindow ? " window" : ""}${isMinimized ? " minimized" : ""}`
}
