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
const http = require('http')

let combinedUrl = (absoluteUrl, relativeUrl) =>  {
	let p = absoluteUrl.lastIndexOf("/");

	if (p < 0) {
		return undefined;
	}

	return new URL(absoluteUrl.substr(0, p + 1) + relativeUrl).href
}

let combinedUrls = (absoluteUrl, relativeUrls) => {
	return relativeUrls.map(relativeUrl => combinedUrl(absoluteUrl, relativeUrl))
}

let combinedPath = (p1, p2) =>  {
    return new URL(p2, "http:/" + p1).toString().substring(6);
};

let iconServiceUrl = (iconNames, size) => {
	if (! Array.isArray(iconNames)) iconNames = [iconNames]
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
let doGet = (service, path, params, changedSince) => {
    return new Promise((resolve, reject) => {
        let req = http.get(url(service, path, params), resp => {
            let data = '';
            resp.setEncoding('utf8');
            resp.on('data', chunk => data += chunk);
            resp.on('end', () => {
                if (resp.statusCode >= 300) {
                    reject(new Error(`Request Failed.\n` + `Status Code: ${resp.statusCode}`));
                } else if (!/^application\/json/.test(resp.headers["content-type"])) {
                    reject(new Error(`Unexpected content-type: ${res.headers["content-type"]}`))
                } else {
                    let json = data === '' ? null : JSON.parse(data);
                    if (typeof json === 'object') {
                        if (Array.isArray(json)) { // A list of resources, then
                            json.forEach(res => adjustIconUrl(res));
                        } else { // A single resource
                            adjustIconUrl(json);
                        }
                        ;
                    }
                    resolve(json);
                }
            });
        });
        req.on('error', e => reject(e));
        req.end();
    });
};

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
            resp.on('data', chunk => { data += chunk });
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

let adjustIconUrl = (res) => {
    res.IconUrl = res.IconUrl ? new URL(res.IconUrl,"http://localhost:7938/" + res.Self.replace(":", "")).toString() :
                  res.IconName ? iconServiceUrl(res.IconName) :
                  undefined;
};

let url = (service, path, params, method) => `http://localhost:7938/${service}${path}${queryString(params)}`;

let opts = (res, method, params) => {
    return {host: 'localhost', port: 7938, path: "/" + res.Self.replace(':', '') + queryString(params), method: method}
};

let queryString = (params) => {
	if (params && Object.keys(params).length > 0) {
        return "?" + Object.keys(params).map(k => k + "=" + params[k]).join("&");
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
		console.log("onOpen, args: ", NW.App.argv);
		WIN.show();
		onOpen && onOpen(args.split(/\s+/));
	})
};

export {nwHide, devtools, NW, nwSetup, iconServiceUrl, doGet, doPost, doDelete}
