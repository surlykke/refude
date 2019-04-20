// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { doGetIfNoneMatch } from '../common/http'
import Axios from 'axios';

Axios.defaults.baseUrl = "http://localhost:7938";

export let getLink = (item, rel) => {
    console.log("getLink:", item, rel)
    if (item && rel && item._links) {
        let link = item._links.find(l => rel === l.rel);
        console.log("getLink returning", link && link.href)
        return link && link.href;
    }
}

export let doGet = (path, dataHandler) => {
    Axios.get(path).then(resp => {
        dataHandler(resp.data)
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

export let monitorUrl = (path, dataHandler) => {
    let etag
    let getIfNoneMatch = () => {
        let headers = { "If-None-Match": etag};
        let validateStatus = status => status === 304 || status < 300 
        Axios.get(path, { headers: headers, validateStatus: validateStatus}).then(resp => {
            if (resp.status < 300) {
                etag = resp.headers.etag
                dataHandler(resp.data)
            }
            setTimeout(getIfNoneMatch, 1000)
        }).catch(err => {
            setTimeout(getIfNoneMatch, 10000) 
        });
    }
    getIfNoneMatch()
}

