// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'

import {storePosition} from "../../common/nw";

class Pinfield extends React.Component {

	constructor(props) {
		super(props)
		this.style = Object.assign({}, props.style)
		Object.assign(this.style, {
			width: "20px",
			height: "20px",
			padding: "2px",
			margin: "0px"
		})
	}

	render = () =>
		<div title="Remember&nbsp;position" style={this.style} onClick={storePosition}>
            <svg viewBox="0 0 24 24">
                <path d="M16,12V4H17V2H7V4H8V12L6,14V16H11.2V22H12.8V16H18V14L16,12M8.8,14L10,12.8V4H14V12.8L15.2,14H8.8Z" />
            </svg>
		</div>
}

export {Pinfield}
