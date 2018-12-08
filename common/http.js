// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
const http = require('http');

const hostPart = "http://locahost:7938";


let iconServiceUrl = (iconNames, size) => {
    if (!Array.isArray(iconNames)) iconNames = [iconNames]
    return "http://localhost:7938/icon-service/icon" +
        "?name=" + iconNames.map(name => encodeURIComponent(name)).join("&name=") +
        "&size=" + (size || 32);
};

export let doGet = options => {
    Object.assign(options, {method: "GET", host: "localhost", port: 7938, path: '/' + options.service + options.path});
    try {
        return new Promise((successHandler, errorHandler) => {
            let req = http.request(options, resp => {
                let data = '';
                resp.setEncoding('utf8');
                resp.on('data', chunk => data += chunk);
                resp.on('end', () => {
                    let o = {
                        status: resp.statusCode,
                        headers: resp.headers,
                        data: data,
                        json: parseAndPostprocess(options.service, data)
                    };
                    if (o.status >= 300) {
                        if (errorHandler) {
                            errorHandler(o);
                        }
                    } else {
                        successHandler(o);
                    }
                });
            });
            req.on('error', e => {
                console.log("req: ", options, " -> ", e);
            });
            req.end();
        })
    }
    catch (e) {
        console.log("Error:", e);
    }
};

export let doGetIfNoneMatch = (service, path, etag) => {
    let options = {
        service: service,
        path: path,
        headers: etag ? {"If-None-Match": `${etag}`} : {}
    };

    return doGet(options);
};

export let doSearch = (service, mimetype, query) => {
    let path = "/search";
    let separator = "?";
    if (mimetype) {
        path += `?type=${encodeURIComponent(mimetype)}`;
        separator = '&';
    }
    if (query) {
        path += `${separator}q=${encodeURIComponent(query)}`;
    }
    let options = {
        service: service,
        path: path
    };

    console.log("Searching with", options)
    return doGet(options);
};

/**
 * We assume no response body
 * @param service
 * @param path
 * @param params
 * @returns {Promise<any>}
 */
export let doPost = (resource, params) => {
    let options = opts(resource, "POST", params);
    return new Promise((resolve, reject) => {
        let req = http.request(options, resp => {
            let data = '';
            resp.setEncoding('utf8');
            resp.on('data', chunk => {
                data += chunk
            });
            resp.on('end', () => {
                if (resp.statusCode >= 300) {
                    reject(new Error(`Request Failed.\nStatus Code: ${resp.statusCode}`));
                } else {
                    resolve(resp);
                }
            });
        });
        req.on('error', e => reject(e));
        req.end();
    });
};

export let doPostPath = (path, params) => {
    return new Promise((resolve, reject) => {
        console.log("Posting against: ", "http://localhost:7938" + path + queryString(params))
        let req = http.request("http://localhost:7938" + path + queryString(params), {method: "POST"}, resp => {
            let data = '';
            resp.setEncoding('utf8');
            resp.on('data', chunk => {
                data += chunk
            });
            resp.on('end', () => {
                if (resp.statusCode >= 300) {
                    reject(new Error(`Request Failed.\nStatus Code: ${resp.statusCode}`));
                } else {
                    resolve(resp);
                }
            });
        });
        req.on('error', e => reject(e));
        req.end();
    });
};

export let doPatch = (resource, body) => {
    return new Promise((resolve, reject) => {
        let req = http.request(opts(resource, "PATCH"), resp => {
            let data = '';
            resp.setEncoding('utf8');
            resp.on('data', chunk => {
                data += chunk
            });
            resp.on('end', () => {
                if (resp.statusCode >= 300) {
                    reject(new Error(`Request Failed.\nStatus Code: ${resp.statusCode}`));
                } else {
                    resolve(resp);
                }
            });
        });
        req.on('error', e => reject(e));
        req.write(JSON.stringify(body));
        req.end();
    });
};

export let doDelete = (resource) => {
    return new Promise((resolve, reject) => {
        let req = http.request(opts(resource, "DELETE"), resp => {
            resp.on('end', () => {
                if (resp.statusCode >= 300) {
                    reject(new Error(`Request Failed.\nStatus Code: ${resp.statusCode}`));
                } else {
                    resolve(resp)
                }
            });
        });
        req.on('error', e => reject(e));
        req.end();
    });
};

let parseAndPostprocess = (service, data) => {
    if (data) {
        try {
            let json = JSON.parse(data);
            if (Array.isArray(json)) {
                json.forEach(resource => {
                    resource._self = "/" + service + resource._self;
                    resource.IconUrl = iconServiceUrl(resource.IconName);
                });
            } else {
                json._self = "/" + service + json._self;
                json.IconUrl = iconServiceUrl(json.IconName);
            }
            return json;
        }
        catch (e) {
            return undefined;
        }
    } else {
        return undefined;
    }
}

let opts = (res, method, params) => {
    return {host: 'localhost', port: 7938, path: encodeURI(res._self) + queryString(params), method: method}
};

let queryString = (params) => {
    if (params && Object.keys(params).length > 0) {
        return "?" + Object.keys(params).map(k => k + "=" + encodeURIComponent(params[k])).join("&");
    } else {
        return "";
    }
};


