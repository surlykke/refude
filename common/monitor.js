import Axios from "axios";
Axios.defaults.baseURL = 'http://localhost:7938'

// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
export let getLink = (item, rel) => {
    if (item && rel && item._links) {
        let link = item._links.find(l => rel === l.rel);
        return link && link.href;
    }
}

export let doGet = (path, dataHandler) => {
    Axios.get(path).then(resp => {
        dataHandler(resp)
    }).catch(err => {
        console.log("GET", path, "got:", err)
    })
}

export let doPost = (path, successHandler) => {
    Axios.post(path).then(resp => {
        successHandler && successHandler(resp);
    }).catch(err => {
        console.log("POST", path, "got:", err)
    });
}

export let doPatch = (path, body, successHandler) => {
    Axios.patch(path, body).then(resp => {
        successHandler && successHandler(resp)
    }).catch(err => {
        console.log("PATCH", path, "got:", err)
    });
};

// Call with a path or a function producing a path
export let monitorUrl = (path, dataHandler) => {
    let etag
    let getIfNoneMatch = () => {
        let actualPath = typeof(path) === "function" ? path() : path
        let headers = { "If-None-Match": etag};
        let validateStatus = status => status === 304 || status < 300 
        Axios.get(actualPath, { headers: headers, validateStatus: validateStatus}).then(resp => {
            if (resp.status < 300) {
                etag = resp.headers.etag
                dataHandler(resp)
            }
            setTimeout(getIfNoneMatch, 1000)
        }).catch(err => {
            setTimeout(getIfNoneMatch, 10000) 
        });
    }
    getIfNoneMatch()
}

