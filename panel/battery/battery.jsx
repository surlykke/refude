/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
import React from 'react'
import {MakeServiceProxy} from '../../common/service-proxy'

const powerProxy = MakeServiceProxy("http://localhost:7938/power-service", "/devices/")

const displayDeviceUrl = "http://localhost:7938/power-service/devices/DisplayDevice"

class Battery extends React.Component {

	constructor(props) {
		super(props)
		this.state = { charge: "?", style: {color: "red"}}
	}

	componentDidMount = () => {
		powerProxy.subscribe(this.updateBattery)
	}

	updateBattery = url => {
		if (url === displayDeviceUrl) {
			let displayDevice = powerProxy.get(displayDeviceUrl)
			if (displayDevice) {
				this.setState({charge: displayDevice.Percentage})
				if (["Charging", "Fully charged"].includes(displayDevice.State)) {
					this.setState({style: {fontWeight: "bold"}})
				}
				else if (displayDevice.Percentage < 20) {
					this.setState({style: {color: "red"}})
				}
				else {
					this.setState({style: {}})
				}
			}
			else {
				this.setState({charge: "?", level: "low", charging: false})
			}
		}
	}

	render = () =>
        <div className="panel-plugin battery" style={this.state.style}>
			{this.state.charge}%
        </div>
}

export {Battery}
