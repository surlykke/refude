// Copyright (c) 2015, 2016, 2017 Christian Surlykke
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
            etag = undefined
            errorHandler && errorHandler(err)
            setTimeout(getIfNoneMatch, 10000)
        });
    }
    getIfNoneMatch()
}

export let getUrl = (path, handler) => {
    console.log("Getting url:", `http://localhost:7938${path}`)
    Axios.get(`http://localhost:7938${path}`).then(resp => handler(resp)).catch(err => console.error(err))
}

export let postUrl = (path, handler) => {
    Axios.post(`http://localhost:7938${path}`).then(resp => {
        handler && handler(resp)
    }).catch(err => console.error(err))
}


