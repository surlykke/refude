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
    	let runClock = () => {
			let now = new Date()
	        this.setState({time: now.toLocaleTimeString()});
			// Update just after next turn of second..
			setTimeout(runClock, 1000 - now.getMilliseconds() + 1);
	    };
		runClock()
	}

	render = () =>
		<div id="clock" className="panel-plugin clock">{this.state.time}</div>
}

export {Clock}
