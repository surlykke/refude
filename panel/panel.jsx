/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
import React from 'react'
import {render} from 'react-dom'
import {nwSetup} from '../common/utils'
import {Clock} from './clock/clock'
import {Battery} from './battery/battery'
import {NotifierItems} from './notifieritems/notifieritems'
import {HideButton} from './hidebutton/hidebutton'
import {DragField} from './dragfield/dragfield'
import {Notifications} from './notifications/notifications'

const Window  = window.require('nw.gui').Window.get()

class Panel extends React.Component {

	constructor(props) {
		super(props)

		nwSetup((argv) => {})
	}

	componentDidMount = () => {
		this.adjustSize()
	}

	adjustSize = () => {
		setTimeout(
			() => {
				let {width, height} = this.content.getBoundingClientRect()
				Window.resizeTo(Math.round(width) - 1, Math.round(height))
			},
			10
		)
	}

	render = () =>
		<div className="wrapper">
	        <div className="content" id="content" ref={div => {this.content = div}}>
				<Clock/>
				<Battery onUpdated={this.adjustSize}/>
				<NotifierItems onUpdated={this.adjustSize}/>
				<HideButton/>
				<DragField/>
				<Notifications onUpdated={this.adjustSize}/>
	        </div>
		</div>
	}

render(
	<Panel/>,
	document.getElementById('root')
);
