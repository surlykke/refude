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
import React from 'react'

class Clock extends React.Component {

	constructor(props) {
		super(props)
		this.state = {time: "--:--:--"}
	}

	componentDidMount = () => {
    	let update = () => {
			let now = new Date()
	        this.setState({time: now.toLocaleTimeString()});
			// Update just after next turn of second..
            setTimeout(update, 1000 - now.getMilliseconds() + 1);
	    };
		update()
	}

	render = () =>
		<div id="clock" style={this.props.style}>{this.state.time}</div>
}

export {Clock}