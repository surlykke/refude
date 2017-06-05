import React from 'react'
//import {nwHide, nwShow} from '../common/utils'

const Window  = window.require('nw.gui').Window.get()

class HideButton extends React.Component {

	constructor(props) {
		super(props)
	}

	hide5s = (event) => {
		Window.minimize()
		setTimeout(() => {Window.restore()}, 5000)
	}

	render = () => <div className="panel-plugin hidebutton" onClick={this.hide5s}/>
}

export {HideButton}
