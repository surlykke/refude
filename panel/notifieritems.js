// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import { getUrl, postUrl, iconUrl, monitorSSE } from "../common/monitor";
import { publish } from "../common/utils";

export class NotifierItem extends React.Component {
    constructor(props) {
        super(props);
        this.state = { item: props.item };
    }

    componentWillReceiveProps = (props) => {
        this.setState({ item: props.item });
    };


    render = () => {
        let showMenu = (event) => {
            console.log("Into showMenu, this.state.item.Data.Menu:", this.state.item.Data.Menu)
            event.preventDefault()
            let menuSelf
            let buildMenu = entries => {
                console.log("Into buildMenu, entries:", entries)
                let menu = new nw.Menu()
                entries.forEach(jsonMenuItem => {
                    let menuItem = new nw.MenuItem({
                        type: jsonMenuItem.Type === "separator" ? "separator" :
                            jsonMenuItem.ToggleType === "checkmark" ? "checkbox" :
                                jsonMenuItem.ToggleType === "radio" ? "checkbox" :
                                    "normal",
                        label: (jsonMenuItem.Label || "").replace(/_([^_])/g, "$1"),
                        checked: jsonMenuItem.ToggleState === 1
                    })
                    if (jsonMenuItem.SubEntries) {
                        menuItem.submenu = buildMenu(jsonMenuItem.SubEntries)
                    } else if (menuItem.type === "normal" || menuItem.type === "checkbox") {
                        menuItem.click = () => {
                            console.log("Post:", menuSelf + '?id=' + jsonMenuItem.Id)
                            postUrl(menuSelf + '?id=' + jsonMenuItem.Id, resp => {
                                console.log("resp:", resp)
                            })
                        }
                    }
                    menu.append(menuItem)
                })
                
                return menu
            }

            if (this.state.item.Data.Menu) {
                getUrl(this.state.item.Data.Menu, resp => {
                    let menu = resp.data 
                    menuSelf = menu.Self
                    let m = buildMenu(menu.Data)
                    m.popup(event.clientX, event.clientY)
                })
            }
        }

        let getXY = (event) => {
            return {
                x: Math.round(event.view.devicePixelRatio * event.screenX),
                y: Math.round(event.view.devicePixelRatio * event.screenY)
            }
        }

        let onClick = (event) => {
            event.persist()
            event.preventDefault()

            let { x, y } = getXY(event)
            if (event.button === 0) {
                let url = this.state.item.Self + '?action=Activate&x=' + x + '&y=' + y;
                console.log("POST against", url);
                postUrl(url);
            } else if (event.button === 1) {
                console.log("POST against", this.state.item.Self + '?action=SecondaryActivate&x=' + x + '&y=' + y)
                postUrl(this.state.item.Self + '?action=middle&x=' + x + '&y=' + y);
            }
        }

        let onRightClick = (event) => {
            event.persist()
            event.preventDefault()
            console.log("onRightClick on", this.state.item)
            if (this.state.item.Data.Menu) {
                showMenu(event)
            } else {
                let { x, y } = getXY(event)
                console.log("POST against", this.state.item.Self + '?action=SecondaryActivate&x=' + x + '&y=' + y)
                postUrl(this.state.item.Self + '?action=ContextMenu&x=' + x + '&y=' + y);
            }
        }

        return this.state.item ?
            <img src={iconUrl(this.state.item.IconName)} alt="" height="14px" width="14px"
                style={{ paddingRight: "5px" }} onClick={onClick} onContextMenu={onRightClick} /> :
            null
    }
}


export class NotifierItems extends React.Component {
    constructor(props) {
        super(props)
        this.state = { items: [] };
        this.style = Object.assign({}, props.style);
        this.style.margin = "0px";
        monitorSSE("status_item", this.getItems, this.getItems, () => {this.setState({items: []})})
    }

    componentDidUpdate = () => {
        publish("componentUpdated");
    };

    getItems = () => {
        getUrl("/items", resp => this.setState({ items: resp.data}));
    };


    render = () => {
        return <div style={this.style}>
            {this.state.items.map(item => (<NotifierItem key={item.Self} item={item} />))}
        </div>
    }

}

