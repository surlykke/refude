// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import {div, img} from './utils.js'

export let updateClock = () => {
	let [now, clockDiv] = [new Date(),document.getElementById("clock")]
	if (clockDiv) clockDiv.innerText = now.toLocaleString()
	setTimeout(updateClock, 1000 - now.getMilliseconds() + 2);
};

export let panel = (trayItems) => {
	return div({className: "panel"}, 
				div({className: "clock",id: "clock"}, new Date().toLocaleString()), 
				...trayItems.map(trayItem => div(null, img({src: trayItem.icon, alt: "", height:"20px", width:"20px"}))))
}
