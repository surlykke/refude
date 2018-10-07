// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import {render} from 'react-dom'
import {doGetIfNoneMatch, doPost} from '../../common/http'
import {monitorResources} from "../common/monitor";

let NotifierItem = (props) => {

    let showMenu = (event) => {
        let buildMenu = (jsonMenu) => {
            let menu = new nw.Menu()
            jsonMenu.forEach(jsonMenuItem => {

                let menuItem = new nw.MenuItem({
                    type: jsonMenuItem.Type === "separator" ? "separator" :
                        jsonMenuItem.ToggleType === "checkmark" ? "checkbox" :
                            jsonMenuItem.ToggleType === "radio" ? "checkbox" :
                                "normal",
                    label: (jsonMenuItem.Label || "").replace(/_([^_])/g, "$1"),
                    checked: jsonMenuItem.ToggleState === 1
                })
                if (jsonMenuItem.SubMenus) {
                    menuItem.submenu = buildMenu(jsonMenuItem.SubMenus)
                } else if (menuItem.type === "normal" || menuItem.type === "checkbox") {
                    menuItem.click = () => {
                        doPost(props.item, {action: "menu", id: jsonMenuItem.Id});
                    }
                }
                menu.append(menuItem)
            })
            return menu
        };

        buildMenu(props.item.Menu).popup(event.clientX, event.clientY)
    }

    let getXY = (event) => {
        return {
            x: Math.round(event.view.devicePixelRatio * event.screenX),
            y: Math.round(event.view.devicePixelRatio * event.screenY)
        }
    }

    let onClick = (event) => {
        event.persist()
        let {x, y} = getXY(event)
        if (event.button === 0) {
            doPost(props.item, {action: "left", x: x, y: y});
        } else if (event.button === 1) {
            doPost(props.item, {action: "middle", x: x, y: y});
        }
        event.preventDefault()
    }

    let onRightClick = (event) => {
        event.persist()
        if (props.item.Menu) {
            showMenu(event)
        } else {
            doPost(...props.item._self.split(":"), {action: "right", x: x, y: y});
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
        monitorResources("statusnotifier-service", "application/vnd.org.refude.statusnotifieritem+json", items => this.setState({items: items}));
    };

    componentDidUpdate = () => {
        console.log("NotifierItems did update", new Date().getMilliseconds());
        this.onUpdated();
    };

    render = () =>
        <div style={this.style}>
            {this.state.items.map((item) => (<NotifierItem key={item._self} item={item}/>))}
        </div>
}

export {NotifierItems}
