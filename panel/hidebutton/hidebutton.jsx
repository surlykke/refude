// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
//import {nwHide, nwShow} from '../common/utils'

const Window  = window.require('nw.gui').Window.get()

class HideButton extends React.Component {

	constructor(props) {
		super(props)
	}

	hide5s = (event) => {
		Window.minimize()
		setTimeout(() => {Window.restore()}, 5000)
	}

	render = () =>
		<div className="panel-plugin hidebutton" onClick={this.hide5s}>
			<svg viewBox="0 0 100 100" >
				<g fillOpacity="0" strokeWidth="10" stroke="black">
				    <rect x="5" y="5" width="90" height="90" />
					<rect x="40" y="40" width="53" height="53"/>
				</g>
			</svg>
		</div>
}

export {HideButton}
