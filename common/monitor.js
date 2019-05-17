// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//

import Axios from "axios";
Axios.defaults.baseURL = 'http://localhost:7938'

// Call with a path or a function producing a path
export let monitorUrl = (path, dataHandler, errorHandler) => {
    let etag
    path = path + "?longpoll"
    let getIfNoneMatch = () => {
        let headers = { "If-None-Match": etag};
        let validateStatus = status => status === 404 || status === 304 || status < 300 
        Axios.get(path, { headers: headers, validateStatus: validateStatus}).then(resp => {
            if (resp.status == 404) {
                // resource must have gone away - stop monitoring
                return
            } else if (resp.status < 300) {
                etag = resp.headers.etag
                dataHandler(resp)
            }
            setTimeout(getIfNoneMatch, 0)
        }).catch(err => {
            // Something more wicked - could be RefudeServices are down. We wait a bit befor trying again
            etag = undefined
            errorHandler && errorHandler(err) 
            setTimeout(getIfNoneMatch, 10000) 
        });
    }
    getIfNoneMatch()
}

