// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import {render} from 'react-dom'
import {doGet, doPost} from '../../common/utils'

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
						doPost(...props.item.Self.split(":"), {action: "menu", id: jsonMenuItem.Id});
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
            doPost(...props.item.Self.split(":"), {action: "left", x: x, y: y});
		} else if (event.button === 1){
            doPost(...props.item.Self.split(":"), {action: "middle", x: x, y: y});
		}
		event.preventDefault()
	}

	let onRightClick = (event) => {
		event.persist()
		if (props.item.Menu) {
			showMenu(event)
		} else {
            doPost(...props.item.Self.split(":"), {action: "right", x: x, y: y});
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
		this.style = Object.assign({}, props.style)
		this.style.margin = "0px"
	}

	componentDidMount = () => {
        let itemCompare = (i1, i2) => i1.Self.localeCompare(i2.Self); // Just to keep them from flipping around
		let update = () => {
            doGet("statusnotifier-service", "/search").then(items => {
                this.setState({items: items.sort(itemCompare)});
            }).catch().then(setTimeout(update, 1000));
        };
		update();
	};

	componentDidUpdate = () => {
		this.onUpdated();
	};

	render = () =>
		<div style={this.style}>
			{this.state.items.map((item) => (<NotifierItem key={item.Self} item={item} /> ))}
		</div>
}

export {NotifierItems}
