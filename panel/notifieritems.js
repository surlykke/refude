// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import { monitorUrl, doGet, getLink, doPost } from "../common/monitor";
import { publish } from "../common/utils";

export class NotifierItem extends React.Component {
    constructor(props) {
        super(props);
        this.state = {};
    }

    componentDidMount = () => {
        monitorUrl(this.props.path, data => {
            console.log("Got item:", data)
            this.setState({ item: data })
        });
    }

    render = () => {
        console.log("Render item");
        let showMenu = (event) => {
            let buildMenu = (jsonMenu) => {
                console.log("Building menu from", jsonMenu);
                let menu = new nw.Menu()
                jsonMenu.Menu.forEach(jsonMenuItem => {

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
                            doPost(this.state.item._self + '?action=menu&id=' + jsonMenuItem.Id);
                        }
                    }
                    menu.append(menuItem)
                })
                return menu
            };

            let href = getLink(this.state.item, "http://relations.refude.org/sni_menu")
            if (href) {
                doGet(href, data => buildMenu(data).popup(event.clientX, event.clientY));
            }
        };

        let getXY = (event) => {
            return {
                x: Math.round(event.view.devicePixelRatio * event.screenX),
                y: Math.round(event.view.devicePixelRatio * event.screenY)
            }
        }

        let onClick = (event) => {
            event.persist()
            let { x, y } = getXY(event)
            if (event.button === 0) {
                console.log("posting on item:", this.state.item);
                let postUrl = this.state.item._self + '?action=left&x=' + x + '&y=' + y;
                console.log("Against:", postUrl)
                doPost(postUrl);
            } else if (event.button === 1) {
                doPost(this.state.item._self + '?action=middle&x=' + x + '&y=' + y);
            }
            event.preventDefault()
        }

        let getMenu = () => {
            let link = this.state.item && this.state.item.Links.find(link => link.rel === "http://relations.refude.org/sni_menu")
            if (link) {

            }
        }

        let onRightClick = (event) => {
            event.persist()
            let { x, y } = getXY(event)
            showMenu(event)
            event.preventDefault()
        }

        let iconUrl = () => {
            console.log("iconUrl returning", '/icon?name=' + this.state.item.IconName)
            return 'http://localhost:7938/icon?name=' + this.state.item.IconName
        }

        return this.state.item ?
            <img src={iconUrl()} alt="" height="18px" width="18px"
                style={{ paddingRight: "5px" }} onClick={onClick} onContextMenu={onRightClick} /> :
            null
    }
}


export class NotifierItems extends React.Component {
    constructor(props) {
        super(props)
        this.state = { itemPaths: [] };
        this.style = Object.assign({}, props.style);
        this.style.margin = "0px";
    }

    componentDidMount = () => {
        monitorUrl("/items?brief", data => this.setState({ itemPaths: data }));
    };

    componentDidUpdate = () => {
        publish("componentUpdated");
    };

    render = () => {
        console.log("Render item");
        return <div style={this.style}>
            {this.state.itemPaths.map(path => (<NotifierItem key={path} path={path} />))}
        </div>
    }

}

