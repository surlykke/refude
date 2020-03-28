// Copyright (c) Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//

import Axios from "axios";

// Call with a path or a function producing a path
export let monitorUrl = (path, dataHandler, errorHandler) => {
    let etag
    let getIfNoneMatch = () => {
        let headers = { "If-None-Match": etag };
        let validateStatus = status => status === 404 || status === 304 || status < 300
        Axios.get(`http://localhost:7938${path}?longpoll`, { headers: headers, validateStatus: validateStatus }).then(resp => {
            if (resp.status == 404) {
                // resource must have gone away - stop monitoring
                return
            } else if (resp.status < 300) {
                etag = resp.headers.etag
                dataHandler(resp)
            }
            getIfNoneMatch()
        }).catch(err => {
            // Something more wicked - could be RefudeServices are down. We wait a bit befor trying again
            console.log(err)
            etag = undefined
            errorHandler && errorHandler(err)
            setTimeout(getIfNoneMatch, 10000)
        });
    }
    getIfNoneMatch()
}


export let monitorSSE = (eventType, onMessage, onOpen, onError) => {
    let helper = () => {
        let url = eventType ? `http://localhost:7938/events?type=${eventType}` : "http://localhost:7938/events" 
        
        let evtSource = new EventSource(url)
        
		evtSource.onerror = event => {
            onError && onError()
			if (evtSource.readyState === 2) {
				setTimeout(helper, 5000)
			}
		}

        evtSource.onopen = onOpen
        evtSource.onmessage  = onMessage
    }

    helper()
}

export let getUrl = (path, handler) => {
    Axios.get(`http://localhost:7938${path}`).then(resp => handler(resp)).catch(err => console.error(err))
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

export let iconUrl = (iconName) => {
    // TODO make icontheme configurable
    return `http://localhost:7938/icon?name=${iconName}&theme=oxygen`;
}

