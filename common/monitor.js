// Copyright (c) Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//

import Axios from "axios";

export let monitorPath = (path, onMessage, onOpen, onError) => {
    let helper = () => {
        let evtSource = new EventSource(`http://localhost:7938/watch?path=${path}`)

        evtSource.onerror = event => {
            onError && onError()
            if (evtSource.readyState === 2) {
                setTimeout(helper, 5000)
            }
        }

        evtSource.onopen = onOpen
        evtSource.onmessage = onMessage
    }

    helper()
}

export let getUrl = (path, handler, errHandler) => {
    Axios.get(`http://localhost:7938${path}`)
        .then(resp => handler(resp))
        .catch(errHandler || (err => console.error(err)))
}

export let postUrl = (path, handler) => {
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


export let iconUrl = (iconName) => {
    // TODO make icontheme configurable
    return `http://localhost:7938/icon?name=${iconName}&theme=oxygen`;
}

