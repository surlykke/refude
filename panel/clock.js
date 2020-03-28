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


// TODO i18n
const monthNames = ["jan", "feb", "mar", "apr", "maj", "jun", "jul", "aug", "sep", "okt", "nov", "dec"];

class Clock extends React.Component {

	constructor(props) {
		super(props)
		this.style = Object.assign({}, this.props.style);
		this.style.fontSize = "14px"; 
		this.style.verticalAlign = "top";
		this.state = {time: "--:--:--"}
	}
	
	componentDidMount = () => {
    	let update = () => {
			let now = new Date()
	        this.setState({
				// TODO i18n
				clockString: `${now.getUTCDate()}. ${monthNames[now.getUTCMonth()]} ${now.getHours()}:${now.getMinutes().toString().padStart(2, "0")}`
			});
			// Update just after next turn of second..
            setTimeout(update, 60000 - now.getSeconds()*1000, now.getMilliseconds() + 1);
	    };
		update()
	}

	render = () => {
		return <div id="clock" style={this.style}>
					<div>{this.state.clockString}</div>
				</div>
	}
}

export {Clock}
