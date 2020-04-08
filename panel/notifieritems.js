// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import { getUrl, postUrl, iconUrl, monitorSSE } from "../common/monitor";
import { remote } from 'electron'
const {Menu, MenuItem} = remote

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
            event.preventDefault()
            let menuSelf
            
            let clickHandler = (id) => {
                return () => {postUrl(`${menuSelf}?id=${id}`)}
            }
           
            let buildMenu = entries => {
                let menu = new Menu()
                entries.forEach(jsonMenuItem => {
                    let label = (jsonMenuItem.Label || "").replace(/_([^_])/g, "$1")

                    if (jsonMenuItem.SubEntries) {
                        menu.append(new MenuItem({label: label, type: "submenu", submenu: buildMenu(jsonMenuItem.SubEntries)}))
                    } else if (jsonMenuItem.Type === "separator") {
                        menu.append(new MenuItem({type: "separator"})) 
                    } else if (jsonMenuItem.ToggleType === "checkmark") {
                        menu.append(new MenuItem({label: label, type: "checkbox", click: clickHandler(jsonMenuItem.Id)}))
                    } else if (jsonMenuItem.ToggleType === "radio") {
                        menu.append(new MenuItem({label: label, type: "radio", click: clickHandler(jsonMenuItem.Id)}))
                    } else {
                        menu.append(new MenuItem({label: label, type: "normal", click: clickHandler(jsonMenuItem.Id)}))
                    }
                   
                })
                
                return menu
            }

            if (this.state.item.Data.Menu) {
                getUrl(this.state.item.Data.Menu, resp => {
                    let menu = resp.data 
                    menuSelf = menu.Self
                    let m = buildMenu(menu.Data)
                    m.popup()
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
                postUrl(url);
            } else if (event.button === 1) {
                postUrl(this.state.item.Self + '?action=middle&x=' + x + '&y=' + y);
            }
        }

        let onRightClick = (event) => {
            event.persist()
            event.preventDefault()
            if (this.state.item.Data.Menu) {
                console.log("showMenu")
                showMenu(event)
            } else {
                let { x, y } = getXY(event)
                console.log(postUrl)
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

    getItems = () => {
        getUrl("/items", resp => this.setState({ items: resp.data}));
    };


    render = () => {
        return <div style={this.style}>
            {this.state.items.map(item => (<NotifierItem key={item.Self} item={item} />))}
        </div>
    }

}

