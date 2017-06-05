/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
import React from 'react'
import {render} from 'react-dom'

import {Clock} from './clock/clock'
import {Battery} from './battery/battery'
import {NotifierItems} from './notifieritems/notifieritems'
import {HideButton} from './hidebutton/hidebutton'

class Panel extends React.Component {

	constructor(props) {
		super(props)
	}

	render = () =>
        <div className="content" id="content">
			<Clock/>
			<Battery/>
			<NotifierItems/>
			<HideButton/>
        </div>
}

render(
	<Panel/>,
	document.getElementById('root')
);
