/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
const http = require('http')

let combinedUrl = (absoluteUrl, relativeUrl) =>  {
	if (relativeUrl[0] === "/") {
		return relativeUrl;
	}

	let p = absoluteUrl.lastIndexOf("/");

	if (p < 0) {
		return undefined;
	}

	return absoluteUrl.substr(0, p + 1) + relativeUrl;
}

let combinedUrls = (absoluteUrl, relativeUrls) => {
	return relativeUrls.map(relativeUrl => combinedUrl(absoluteUrl, relativeUrl))
}


let iconServiceUrl = (iconName, size) => {
    return "http://localhost:7938/icon-service/icon?name=" + iconName + "&size=" + (size || 32);
}

const tcpPattern = /http:\/\/(\w*)(:(\d+))?(\/.*)/

let doHttp = (url, method) => { // TODO payload
	let m = tcpPattern.exec(url)

	let opts = {
		host: m[1],
		port: m[3] || 80,
		path: m[4],
		method: method || "GET"
	}

	console.log("doHttp, opts: ", opts)

	return new Promise((resolve, reject) => {
		http.request(opts, resp => {
			let data = ''
			resp.setEncoding('utf8')
			resp.on('data', chunk => { data += chunk })
			resp.on('end', () => {
				try {
					resolve(data === '' ? null : JSON.parse(data))
				}
				catch (e) {
					reject(e)
				}
			})
		})
		.on('error', e => reject(e))
		.end()
	})

}

export {combinedUrl, combinedUrls, iconServiceUrl, doHttp}
