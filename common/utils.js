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

const tcpPattern = /http:\/\/(\w*)(:(\d+))?(\/.*)/

let doHttp = (url, method, payload) => {
	let m = tcpPattern.exec(url)

	let opts = {
		host: m[1],
		port: m[3] || 80,
		path: m[4],
		method: method || "GET"
	}

	return new Promise((resolve, reject) => {
		let req = http.request(opts, resp => {
			let data = ''
			resp.setEncoding('utf8')
			resp.on('data', chunk => { data += chunk })
			resp.on('end', () => {
				if (resp.statusCode > 299) {
					reject(new Error(`Request Failed.\n` + `Status Code: ${resp.statusCode}`));
				}
				else {
					try {
						resolve(data === '' ? null : JSON.parse(data))
					}
					catch (e) {
						reject(e)
					}
				}
			})
		})
		req.on('error', e => reject(e))
		if (payload) {
			req.write(JSON.stringify(payload))
		}
		req.end()
	})

}
let doHttp2 = (service, path, method, payload) => {
	return new Promise((resolve, reject) => {
		let req = http.request({host: "localhost", port: 7938, path: "/" + service + path, method: method || "GET"}, resp => {
			let data = '';
			resp.setEncoding('utf8');
			resp.on('data', chunk => { data += chunk });
			resp.on('end', () => {
				if (resp.statusCode > 299) {
					reject(new Error(`Request Failed.\n` + `Status Code: ${resp.statusCode}`));
				}
				else {
					try {
						resolve(data === '' ? null : JSON.parse(data))
					}
					catch (e) {
						reject(e)
					}
				}
			})
		});
		req.on('error', e => reject(e));
		if (payload) {
			req.write(JSON.stringify(payload));
		}
		req.end();
	})
};

let doPost = (service, path, params) => {
	path = path + queryString(params);
	return doHttp2(service, path, "POST");
};

let doGet = (service, path, params) => {
	path = path + queryString(params);
    return doHttp2(service, path);
}

let doDelete = (service, path, params) => {
    path = path + queryString(params);
    return doHttp2(service, path);
}

let adjustIconUrl = (res) => {
    res.IconUrl = res.IconUrl ? new URL(res.IconUrl,"http://localhost:7938/" + res.Self.replace(":", "")).toString() :
                  res.IconName ? iconServiceUrl(res.IconName) :
                  undefined;
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

export {nwHide, devtools, NW, nwSetup, combinedUrl, combinedUrls, combinedPath, iconServiceUrl, doHttp, doHttp2, doPost, doGet, adjustIconUrl}
