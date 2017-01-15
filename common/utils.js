/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
let combineUrls = function(url, relativeUrl) {
	if (relativeUrl[0] === "/") {
		return relativeUrl;
	}

	let p = url.lastIndexOf("/");
	
	if (p < 0) {
		return undefined;
	}
	
	return url.substr(0, p + 1) + relativeUrl;
}