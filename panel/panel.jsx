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

import {Clock} from './clock/clock'
import {Battery} from './battery/battery'
import {NotifierItems} from './notifieritems/notifieritems'

const gui = window.require('nw.gui')
const Window = gui.Window.get()
class Panel extends React.Component {

	constructor(props) {
		super(props)
	}

	onClick = (event) => {
		Window.minimize()
		setTimeout(() => {Window.restore()}, 5000)
	}

	render = () =>
        <div className="content" onClick={this.onClick}>
			<Clock/>
			<Battery/>
			<NotifierItems/>
        </div>
}



render(
	<Panel/>,
	document.getElementById('root')
);
