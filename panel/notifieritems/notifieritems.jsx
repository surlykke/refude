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

	let showMenu = (event) => {
		((jsonMenu) => {
			let menu = new nw.Menu()
			jsonMenu.forEach(jsonMenuItem => {

				let menuItem = new nw.MenuItem({
					type: jsonMenuItem.Type === "separator" ? "separator" :
					      jsonMenuItem.ToggleType === "checkmark" ? "checkbox" :
					      jsonMenuItem.ToggleType === "radio" ? "checkbox" :
						  "normal",
					label: (jsonMenuItem.Label || "").replace( /_([^_])/g, "$1" ),
					checked: jsonMenuItem.ToggleState === 1
				})
				if (jsonMenuItem.SubMenus) {
					menuItem.submenu = buildMenu(jsonMenuItem.SubMenus)
				} else if (menuItem.type === "normal" || menuItem.type === "checkbox") {
					menuItem.click = () => {
						doHttp(`${props.item.url}?action=menu&id=${jsonMenuItem.Id}`, "POST")
					}
				}
				menu.append(menuItem)
			})
			return menu
		})(props.item.Menu).popup(event.clientX, event.clientY)
	}

	let getXY = (event) => {
		return  {
			x: Math.round(event.view.devicePixelRatio * event.screenX),
			y: Math.round(event.view.devicePixelRatio * event.screenY)
		}
	}

	let onClick = (event) => {
		event.persist()
		let {x,y} = getXY(event)
		if (event.button === 0) {
			doHttp(`${props.item.url}?action=left&x=${x}&y=${y}`, "POST")
		} else if (event.button === 1){
			doHttp(`${props.item.url}?action=middle&x=${x}&y=${y}`, "POST")
		}
		event.preventDefault()
	}

	let onRightClick = (event) => {
		event.persist()
		if (props.item.Menu) {
			showMenu(event)
		} else {
			doHttp(`${props.item.url}?action=right&x=${x}&y=${y}`, "POST")
		}
		event.preventDefault()
	}


	return (<img src={props.item.IconUrl} height="18px" width="18px"
	             style={{paddingRight: "5px"}} onClick={onClick} onContextMenu={onRightClick}/>)
}


class NotifierItems extends React.Component {
	constructor(props) {
		super(props)
		this.state = {items : []}
		this.onUpdated = props.onUpdated
		this.items = MakeCollection("statusnotifier-service", "/items", this.update)
		this.style = Object.assign({}, props.style)
		this.style.margin = "0px"
	}

	componentDidUpdate() {
		this.onUpdated()
	}

	update = () => {
		this.setState({items: this.items.all})
	}

	render = () =>
		<div style={this.style}>
			{this.state.items.map((item) => (<NotifierItem key={item.id} item={item} /> ))}
		</div>
}

export {NotifierItems}
