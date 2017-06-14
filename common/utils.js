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

// -------------------- NW stuff ---------------------
let NW = window.require('nw.gui')
let WIN = NW.Window.get()

let nwHide = () => {
		WIN.hide()
}


let nwSetup = (onShow) => {
	let nwShow = () => {
		WIN.show();
		onShow && onShow()
	}

	NW.App.on("open", (args) => {nwShow()})
}

export {nwHide,  nwSetup, combinedUrl, combinedUrls, iconServiceUrl, doHttp}
