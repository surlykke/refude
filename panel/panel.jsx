/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
import React from 'react'
import {render} from 'react-dom'
import {MakeServiceProxy} from '../common/service-proxy'
import {nwHide, nwShow} from '../common/utils'

const powerProxy = MakeServiceProxy("http://localhost:7938/power-service/devices", "http://localhost:7938/power-service/notify")
const displayDeviceUrl = "http://localhost:7938/power-service/device/DisplayDevice"

class Panel extends React.Component {

	constructor(props) {
		super(props)
		this.state = {time: "--:--:--", charge: "?", level: "low", charging: false}
	}

	componentDidMount = () => {
		powerProxy.subscribe(this.updateBattery)
		this.runClock()
	}

	updateBattery = url => {
		if (url === displayDeviceUrl) {
			let displayDevice = powerProxy.get(displayDeviceUrl)
			if (displayDevice) {
				this.setState({charge: displayDevice.Percentage + "%",
							   level: displayDevice.Percentage < 20 ? "low" : "ok",
				               charging: ["Charging", "Fully charged"].includes(displayDevice.State)
							  })
			}
			else {
				this.setState({charge: "?", level: "low", charging: false})
			}
		}
	}

    runClock = () => {
		let now = new Date()
        this.setState({time: now.toLocaleTimeString()});
        setTimeout(this.runClock, 1000 - now.getMilliseconds() + 1); // Just after next turn of second..
    };

	style = () => {

		let s = this.state.charging ? {fontWeight: "bold"} :
		        this.state.level === "low" ? {color: "red"} :
			    {}
		console.log("state: ", this.state, ", style: ", s)
		return s
	}

	onClick = (event) => {
		nwHide()
		setTimeout(() => {nwShow()}, 5000)
	}

	render = () =>
        <div className="content" onClick={this.onClick}>
			&nbsp;
			<span id="battery" style={this.style()}>{this.state.charge}</span>&nbsp;
			<span id="clock">{this.state.time}</span> &nbsp;
        </div>
}



render(
	<Panel/>,
	document.getElementById('root')
);
