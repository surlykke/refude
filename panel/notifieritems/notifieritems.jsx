// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import {render} from 'react-dom'
import {MakeCollection} from '../../common/resources'
import {doHttp} from '../../common/utils'

let NotifierItem = (props) => {

	let getXY = (event) => {
		return  {
			x: Math.round(event.view.devicePixelRatio * event.screenX),
			y: Math.round(event.view.devicePixelRatio * event.screenY)
		}
	}

	let onClick = (event) => {
		event.persist()
		console.log(event)
		if (event.button === 0) {
			call("Activate", getXY(event))
		} else if (event.button === 1){
			call("SecondaryActivate", getXY(event))
		}
	}

	let onRightClick = (event) => {
		event.persist()
		call("ContextMenu", getXY(event))
		event.preventDefault()
	}

	let call = (method, xy) => {
		let url = props.item.url + `?method=${method}&x=${xy.x}&y=${xy.y}`
		console.log("Posting: ", url)
		doHttp(url, "POST")
	}
	let style = {
		paddingRight: "5px"
	}
	return (<img src={props.item.IconUrl} height="18px" width="18px" style={style} onClick={onClick} onContextMenu={onRightClick}/>)
}


class NotifierItems extends React.Component {
	constructor(props) {
		super(props)
		this.state = {items : []}
		this.onUpdated = props.onUpdated
		console.log("constructor: this.state.items:", this.state.items)
		this.items = MakeCollection("statusnotifier-service", "/items", this.update)
	}

	componentDidUpdate() {
		this.onUpdated()
	}

	update = () => {
		this.setState({items: this.items.all})
	}

	render = () =>
		<div className="panel-plugin notifier-items">
			{this.state.items.map((item) => (<NotifierItem key={item.id} item={item} /> ))}
		</div>
}

export {NotifierItems}
