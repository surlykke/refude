// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
const http = require('http');

const hostPart = "http://locahost:7938";


let iconServiceUrl = (iconNames, size) => {
    if (!Array.isArray(iconNames)) iconNames = [iconNames]
    return "http://localhost:7938/icon-service/icon" +
        "?name=" + iconNames.map(name => encodeURIComponent(name)).join("&name=") +
        "&size=" + (size || 32);
}

/**
 * We assume the responsebody to be json
 * @param service
 * @param path
 * @param params
 * @param changedSince
 * @returns {Promise<any>}
 */
let doGet = (service, path, params, ifNoneMatchEtag) => {
    let options = {
        method: "GET",
        host: "localhost",
        port: 7938,
        path: `/${service}${path}${queryString(params)}`
    }

    if (ifNoneMatchEtag) {
        options.headers = {"If-None-Match": ifNoneMatchEtag};
    }

    return new Promise((resolve, reject) => {
        let req = http.request(options, resp => {
            let data = '';
            resp.setEncoding('utf8');
            resp.on('data', chunk => data += chunk);
            resp.on('end', () => {
                if (resp.statusCode >= 300) {
                    reject(new Error(`Request Failed.\n` + `Status Code: ${resp.statusCode}`));
                } else if (!/^application\/(.*\+)?json/.test(resp.headers["content-type"])) {
                    reject(new Error(`Unexpected content-type: ${resp.headers["content-type"]}`))
                } else {
                    let json = data === '' ? null : JSON.parse(data);
                    if (typeof json === 'object') {
                        if (Array.isArray(json)) { // A list of resources, then
                            json.sort((e1, e2) => (e2.RelevanceHint || 0) - (e1.RelevanceHint || 0))
                            json.forEach(res => adjustUrls(service, res));
                        } else { // A single resource
                            adjustUrls(service, json);
                        }
                    }
                    if (path === "/display") {
                        resolve({json: json, headers: resp.headers});
                    } else {
                        resolve(json)
                    }

                }
            });
        });
        req.on('error', e => {
            console.log("req: ", options);
            reject(e);
        });
        req.end();
    });
};

let doGetH = (query, handler) => {
    let options = {
        method: "GET",
        host: "localhost",
        port: 7938,
        path: `/${query.service}${query.path}`
    }

    if (query.ifNoneMatch) {
        options.headers = {"If-None-Match": query.ifNoneMatch};
    }

    let req = http.request(options, resp => {
        let data = '';
        resp.setEncoding('utf8');
        resp.on('data', chunk => data += chunk);
        resp.on('end', () => {
            if (resp.statusCode >= 300) {
                // We ignore it. Among others, 304 after if-none-match.
            } else if (!/^application\/(.*\+)?json/.test(resp.headers["content-type"])) {
                console.log(`Unexpected content-type: ${resp.headers["content-type"]}`);
            } else {
                let json = data === '' ? null : JSON.parse(data);
                if (typeof json === 'object') {
                    if (Array.isArray(json)) { // A list of resources, then
                        json.sort((e1, e2) => (e2.RelevanceHint || 0) - (e1.RelevanceHint || 0));
                        json.forEach(res => adjustUrls(service, res));
                    } else { // A single resource
                        adjustUrls(query.service, json);
                    }
                }
                handler(json, resp.headers);
            }
        });
    });
    req.on('error', e => console.log("Request error:", e));
    req.end();
}


/**
 * We assume no response body
 * @param service
 * @param path
 * @param params
 * @returns {Promise<any>}
 */
let doPost = (resource, params) => {
    return new Promise((resolve, reject) => {
        let req = http.request(opts(resource, "POST", params), resp => {
            let data = '';
            resp.setEncoding('utf8');
            resp.on('data', chunk => {
                data += chunk
            });
            resp.on('end', () => {
                if (resp.statusCode >= 300) {
                    reject(new Error(`Request Failed.\n` + `Status Code: ${resp.statusCode}`));
                } else {
                    resolve(resp);
                }
            });
        });
        req.on('error', e => reject(e));
        req.end();
    });
};

let doDelete = (resource) => {
    return new Promise((resolve, reject) => {
        let req = http.request(opts(resource, "DELETE"), resp => {
            resp.on('end', () => {
                if (resp.statusCode >= 300) {
                    reject(new Error(`Request Failed.\n` + `Status Code: ${resp.statusCode}`));
                } else {
                    resolve(resp)
                }
            });
        });
        req.on('error', e => reject(e));
        req.end();
    });
};

let adjustUrls = (service, res) => {
    res._self = "/" + service + res._self;  // Proxy should really do this
    res.IconUrl = iconServiceUrl(res.IconName);
};

let opts = (res, method, params) => {
    return {host: 'localhost', port: 7938, path: res._self + queryString(params), method: method}
};

let queryString = (params) => {
    if (params && Object.keys(params).length > 0) {
        return "?" + Object.keys(params).map(k => k + "=" + encodeURIComponent(params[k])).join("&");
    } else {
        return "";
    }
};

// -------------------- NW stuff ---------------------
let NW = window.require('nw.gui');
let WIN = NW.Window.get();

let nwHide = () => {
    WIN.hide();
};

let devtools = () => {
    WIN.showDevTools();
};

let nwSetup = (onOpen) => {
    NW.App.on("open", (args) => {
        WIN.show();
        onOpen && onOpen(args.split(/\s+/));
    })
};

let displayEtag = null;

let watchPos = () => {
    WIN.on('move', (x, y) => {
        console.log("on move, displayEtag:", displayEtag)
        if (displayEtag) {
            localStorage.setItem(displayEtag + ".x", x);
            localStorage.setItem(displayEtag + ".y", y);
        }
    });
};

let adjustPos = () => {
    doGetH({service: "wm-service", path: "/display", ifNoneMatch: displayEtag}, (json, headers) => {
        if (headers && headers.etag) {
            let x = localStorage.getItem(headers.etag + ".x");
            let y = localStorage.getItem(headers.etag + ".y");
            if (x && y) {
                WIN.moveTo(parseInt(x), parseInt(y));
            }
            displayEtag = headers.etag;
            console.log("setting displayEtag:", displayEtag)
        }
    });
};


export {nwHide, devtools, NW, nwSetup, iconServiceUrl, doGet, doGetH, doPost, doDelete, watchPos, adjustPos}
