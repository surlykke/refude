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
		let buildMenu = (jsonMenu) => {
			let menu = new nw.Menu()
			jsonMenu.forEach(jsonMenuItem => {
				let menuItem = new nw.MenuItem({
					type: jsonMenuItem.Type === "separator" ? "separator" : "normal",
					label: (jsonMenuItem.Label || "").replace( /_([^_])/g, "$1" )
				})
				if (jsonMenuItem.SubMenus) {
					menuItem.submenu = buildMenu(jsonMenuItem.SubMenus)
				} else if (menuItem.type === "normal") {
					menuItem.click = () => {
						let url = props.item.url + `?action=click&id=${jsonMenuItem.Id}`
						doHttp(url, "POST")
					}
				}
				menu.append(menuItem)
			})
			return menu
		}
		let menu = buildMenu(props.item.Menu)

		if (menu.items.length === 1 && menu.items[0].type === "normal" && !menu.items[0].submenu) {
			// So the entire menu consists of one item. We don't bother showing it, just activete it directly.
			menu.items[0].click()
		} else {
			menu.popup(event.clientX, event.clientY)
		}
	}




	let getXY = (event) => {
		return  {
			x: Math.round(event.view.devicePixelRatio * event.screenX),
			y: Math.round(event.view.devicePixelRatio * event.screenY)
		}
	}

	let onClick = (event) => {
		event.persist()
		if (props.item.Menu) {
			showMenu(event)
		} else if (event.button === 0) {
			call("Activate", getXY(event))
		} else if (event.button === 1){
			call("SecondaryActivate", getXY(event))
		}
		event.preventDefault()
	}

	let onRightClick = (event) => {
		event.persist()
		if (props.item.Menu) {
			showMenu(event)
		} else {
			call("ContextMenu", getXY(event))
		}
		event.preventDefault()
	}

	let call = (method, xy) => {
		let url = props.item.url + `?method=${method}&x=${xy.x}&y=${xy.y}`
		console.log("Posting: ", url)
		doHttp(url, "POST")
	}

	console.log("item: ", props.item)

	return (<img src={props.item.IconUrl} height="18px" width="18px"
	             style={{paddingRight: "5px"}} onClick={onClick} onContextMenu={onRightClick}/>)
}


class NotifierItems extends React.Component {
	constructor(props) {
		super(props)
		this.state = {items : []}
		this.onUpdated = props.onUpdated
		console.log("constructor: this.state.items:", this.state.items)
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
