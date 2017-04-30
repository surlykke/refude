/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

// -------------------- NW stuff ---------------------
let NW = window.require('nw.gui')
let WIN = NW.Window.get()

let nwHide = () => {
		WIN.hide()
}

let nwSetup = () => {
	NW.App.on("open", (args) => {
			console.log("Opened", args)
			WIN.show();
	})
}

export {nwHide, nwSetup}
