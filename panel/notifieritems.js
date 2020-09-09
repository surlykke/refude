// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import { getUrl, postUrl, findLink, path2Url } from "../common/monitor";
import { remote, ipcRenderer } from 'electron'
const {Menu, MenuItem} = remote
import './Panel.css'

export class NotifierItem extends React.Component {
    constructor(props) {
        super(props);
        this.state = { };
        this.getItem()
    }

    componentDidMount = () => {
        ipcRenderer.on("sseopen", this.getItem) 
        ipcRenderer.on(this.props.link.href, this.getItem)
        this.getItem()
    }

    componentWillUnmount = () => {
        ipcRenderer.removeListener("sseopen", this.getItem) 
        ipcRenderer.removeListener(this.props.link.href, this.getItem)
    }


    getItem = () => {
        getUrl(this.props.link.href, resp => this.setState({item: resp.data, self: findLink(resp.data, "self")}))
    }
   

    render = () => {
        let showMenu = (event) => {
            event.preventDefault()
            let menuSelf
            
            let clickHandler = (id) => {
                return () => {postUrl(`${this.state.item.Menu}?id=${id}`)}
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

            if (this.state.item.Menu) {
                getUrl(this.state.item.Menu, resp => {
                    let m = buildMenu(resp.data.Entries)
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
            let self = findLink(this.state.item, "self").href
            if (event.button === 0) {
                postUrl(self + '?action=Activate&x=' + x + '&y=' + y);
            } else if (event.button === 1) {
                postUrl(self + '?action=middle&x=' + x + '&y=' + y);
            }
        }

        let onRightClick = (event) => {
            event.persist()
            event.preventDefault()
            if (this.state.item.Menu) {
                showMenu(event)
            } else {
                let { x, y } = getXY(event)
                postUrl(this.state.item.Self + '?action=ContextMenu&x=' + x + '&y=' + y);
            }
        }

        return this.state.item ?
            <img src={path2Url(this.state.self.icon)} alt="" height="14px" width="14px"
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
        ipcRenderer.on("/items", this.getItems) 
        ipcRenderer.on("sseopen", this.getItems)
        ipcRenderer.on("sseerror", this.error)
        this.getItems()
    }
  
    componentWillUnmount = () => {
        ipcRenderer.removeListener("/items", this.getItems) 
        ipcRenderer.removeListener("sseopen", this.getItems)
        ipcRenderer.removeListener("sseerror", this.error)
    }
   
    getItems = () => getUrl("/items", resp => this.setState({ links: resp.data}))

    error = () => this.setState({links: undefined})

    render = () => {
        return this.state.links ? <div className="plugin">
            {this.state.links.map(link => (<NotifierItem key={link.href} link={link} />))}
        </div> : null
    }

}

