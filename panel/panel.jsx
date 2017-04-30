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

const powerProxy = MakeServiceProxy("http://localhost:7938/power-service/devices", "http://localhost:7938/power-service/notify")
const displayDeviceUrl = "http://localhost:7938/power-service/device/DisplayDevice"

class Panel extends React.Component {

	constructor(props) {
		super(props)
		this.state = {time: "--:--:--", charge: 0, state: "?"}
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
				               charging: ["Charging", "Fully charged"].includes(displayDevice.State) ? "yes" :
				               			 ["Discharging", "Empty"].includes(displayDevice.State) ? "no":
										  "?"
							  })
			}
		}
	}

    runClock = () => {
		let now = new Date()
        this.setState({time: now.toLocaleTimeString()});
        setTimeout(this.runClock, 1000 - now.getMilliseconds() + 1); // Just after next turn of second..
    };

	style = () => {
		return this.state.charging === "yes" ? {fontWeight: "bold"}:
		       this.state.charge < 20 ? {color: "red"}:
			                            {}
	}

	render = () =>
        <div className="content">
			&nbsp;
			<span id="battery" style={this.style()}>{this.state.charge}</span>&nbsp;
			<span id="clock">{this.state.time}</span> &nbsp;
        </div>
}



render(
	<Panel/>,
	document.getElementById('root')
);
