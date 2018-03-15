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
import {MakeCollection} from '../../common/resources'

class Battery extends React.Component {

	constructor(props) {
		super(props)
		this.onUpdated = props.onUpdated
		this.state = {data: []}
	}

	componentDidMount = () => {
		this.devices = MakeCollection("power-service", "/devices", this.update, dev => dev.Type === 'Battery' && !dev.DisplayDevice)
	}

	update = () => {
		console.log("devices: ", this.devices);
		console.log("devices.filtered: ", this.devices.filtered)
		this.setState({data: this.devices.filtered.sort((d1,d2) => d1.NativePath.localeCompare(d2.NativePath)).map(b => {
			let charging = ["Charging", "Fully charged"].includes(b.State) 
			return {
				style: { 
					color: (charging || b.Percentage >= 20 ? 'black' : 'red'),
					fontWeight: charging ? 'bold' : 'normal',
					marginRight: '0.4em'
				},
				percentage: Math.floor(b.Percentage + 0.5),
				path: b.NativePath
			}
		})})
		console.log("this.state.data:", this.state.data)
	}


	render = () => {
		return <div style={this.props.style}>{this.state.data.map(d => (<span style={d.style} key={d.path}>{d.percentage}%</span>))}</div>
	}
}

export {Battery}
