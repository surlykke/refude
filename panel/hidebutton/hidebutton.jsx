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
		this.style = Object.assign({}, props.style)
		Object.assign(this.style, {
			width: "16px",
			height: "16px",
			padding: "2px",
			marginRight: "2px"
		})
	}

	hide5s = (event) => {
		Window.minimize()
		setTimeout(() => {Window.restore()}, 5000)
	}

	render = () =>
		<div style={this.style} onClick={this.hide5s}>
			<svg viewBox="0 0 100 100" >
				<g fillOpacity="0" strokeWidth="10" stroke="black">
				    <rect x="10" y="10" width="80" height="80" />
					<rect x="43" y="43" width="47" height="47"/>
				</g>
			</svg>
		</div>
}

export {HideButton}
