// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import { monitorUrl, getUrl, postUrl, iconUrl } from "../common/monitor";
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
            console.log("Into showMenu, this.state.item.Menu:", this.state.item.Menu)
            event.preventDefault()
            let buildMenu = jsonMenu => {
                let menu = new nw.Menu()
                jsonMenu.Entries.forEach(jsonMenuItem => {
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
                            console.log("Post:", jsonMenu._self + '?id=' + jsonMenuItem.Id)
                            postUrl(jsonMenu._self + '?id=' + jsonMenuItem.Id, resp => {
                                console.log("resp:", resp)
                            })
                        }
                    }
                    menu.append(menuItem)
                })
                
                return menu
            }

            if (this.state.item.Menu) {
                getUrl(this.state.item.Menu, resp => {
                    let m = buildMenu(resp.data)
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
                let url = this.state.item._self + '?action=left&x=' + x + '&y=' + y;
                console.log("POST against", url);
                postUrl(url);
            } else if (event.button === 1) {
                console.log("POST against", this.state.item._self + '?action=middle&x=' + x + '&y=' + y)
                postUrl(this.state.item._self + '?action=middle&x=' + x + '&y=' + y);
            }
        }

        let onRightClick = (event) => {
            event.persist()
            let { x, y } = getXY(event)
            showMenu(event)
            event.preventDefault()
        }


        return this.state.item ?
            <img src={iconUrl(this.state.item.IconName)} alt="" height="18px" width="18px"
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
    }

    componentDidMount = () => {
        monitorUrl("/items", resp => this.setState({ items: resp.data }), err => this.setState({ items: [] }));
    };

    componentDidUpdate = () => {
        publish("componentUpdated");
    };

    render = () => {
        return <div style={this.style}>
            {this.state.items.map(item => (<NotifierItem key={item._self} item={item} />))}
        </div>
    }

}

