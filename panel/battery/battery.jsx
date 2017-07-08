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
import {MakeResource} from '../../common/resources'

class Battery extends React.Component {

	constructor(props) {
		super(props)
		this.onUpdated = props.onUpdated
		this.state = {}
	}

	componentDidMount = () => {
		this.battery = MakeResource("power-service", "/devices/DisplayDevice", this.update)
	}

	update = () => {
		// status: 1: warning, 2: discharging, 3: charging/fully charged
		let status = ["Charging", "Fully charged"].includes(this.battery.State) ? 3 :
		             this.battery.Percentage >= 20 ? 2 :
					 1
		this.setState({charge: this.battery.Percentage || "? ", status: status})
		this.onUpdated()
	}

	render = () => {
		let style = this.state.status === 3 ? {fontWeight: "bold"} :
		            this.state.status === 1 ? {color: "red"} :
					{}
		return <div className="panel-plugin battery" style={style}>
					{this.state.charge}%
        	   </div>
	}
}

export {Battery}
