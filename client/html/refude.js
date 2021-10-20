/*
 * Copyright (c) Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

import {div, materialIcon, frag, getJson, doPost, doDelete} from './utils.js'
import { clock } from './clock.js'
import { notifierItem } from './notifieritem.js'
import { battery } from './battery.js'
import {resourceHead} from './resource.js'
import {links} from './link.js'

const startUrl = "/search/desktop"

export class Refude extends React.Component {
    
    constructor(props) {
        console.log("Refude constructor")
        super(props)
        this.history = []
        this.resourceUrl = startUrl 
        this.state = { term: "", itemlist: []}
        this.watchSse()
    }

    componentDidMount = () => {
        document.addEventListener("keydown", this.onKeyDown)
    };

    watchSse = () => {
        let evtSource = new EventSource("http://localhost:7938/watch")

        evtSource.onopen = () => {
            this.getResource()
            this.getDisplayDevice()
            this.getItemlist()
            this.getFlash()
        }

        evtSource.onerror = event => {
            ipcRenderer.send("sseerror")
            if (evtSource.readyState === 2) {
                setTimeout(watchSse, 5000)
            }
        }

        evtSource.onmessage = event => {
            console.log("watch message:", event.data)
            if (this.resourceUrl === event.data) {
                this.getResource()
            } else if ("/device/DisplayDevice" === event.data) {
                this.getDisplayDevice()
            } else if ("/item/list" === event.data) {
                this.getItemlist()
            } else if ("/notification/list" === event.data) {
                this.getFlash()
            } 
        }
    }

    getDisplayDevice = () => {
        fetch("http://localhost:7938/device/DisplayDevice")
            .then(resp => resp.json())
            .then(
                json => this.setState({displayDevice: json.data}),
                error => this.setState({displayDevice: undefined})
            )
    }

    getFlash = () => {
        /* FIXME */
    }

    getItemlist = () => {
        fetch("http://localhost:7938/item/list")
            .then(resp => resp.json())
            .then( 
                json => {this.setState({itemlist: json.data})},
                error => {this.setState({itemlist: []})}
            )
    }

    getResource = () => {
        let href = `${this.resourceUrl}?term=${this.state.term}`
        fetch(href)
            .then(resp => resp.json())
            .then(
                json => {this.setState({resource: json})},
                error => {this.setState({ resource: undefined })}
            )
    }

    select = link => {
        this.currentLink = link
    }

    activate = (link, dismiss) => {
        if (link.rel === "org.refude.defaultaction" || link.rel === "org.refude.action") {
            doPost(link.href).then(() => dismiss && this.dismiss())
        } else if (link.rel === "org.refude.delete") {
            doDelete(link.href).delete(path).then(() => dismiss && this.dismiss())
        } else if (link.rel === "related") {
            getJson(link.href)
                .then(json => {
                    if (json?._links) {
                        let dfAct = json._links.find(l => l.rel === "org.refude.defaultaction")
                        if (dfAct) {
                            doPost(dfAct.href).then(() => dismiss && this.dismiss())
                        }
                    }
                })
        }
    }

    delete = (link, dismiss) => {
        if (link.rel === "org.refude.delete") {
            doDelete(link.href).delete(path).then(() => dismiss && this.dismiss())
        } else if (link.rel === "related") {
            getJson(link.href)
                .then(json => {
                    if (json?._links) {
                        let deleteLink = json._links.find(l => l.rel === "org.refude.delete")
                        if (deleteLink) {
                            doDelete(deleteLink.href).then(() => dismiss && this.dismiss())
                        }
                    }
                })
        }
    }

    dismiss = () => {
        console.log("Posting:", "http://localhost:7938/client/dismiss")
        this.history = []
        this.resourceUrl = startUrl 
        this.setState({ term: "", itemlist: [] })
 
        doPost("http://localhost:7938/client/dismiss")
    }

    navigateTo = href => {
        this.history.unshift(this.resourceUrl)
        this.resourceUrl = href
        this.setState({ term: "" }, this.getResource)
    }

    navigateBack = () => {
        this.resourceUrl = this.history.shift() || this.resourceUrl
        this.setState({ term: "" }, this.getResource)
    }

    onKeyDown = (event) => {
        let { key, keyCode, ctrlKey, altKey, shiftKey, metaKey } = event;
        if (key === "ArrowRight" || key === "l" && ctrlKey) {
            this.currentLink.rel === "related" && this.navigateTo(this.currentLink.href)
        } else if (key === "ArrowLeft" || key === "h" && ctrlKey) {
            this.navigateBack()
        } else if (key === "ArrowUp" || key === "k" && ctrlKey || key === 'Tab' && shiftKey && !ctrlKey && !altKey) {
            this.move(true)
        } else if (key === "ArrowDown" || key === "j" && ctrlKey || key === 'Tab' && !shiftKey && !ctrlKey && !altKey) {
            this.move()
        } else if (key === "Enter") {
            this.activate(this.currentLink, !ctrlKey)
        } else if (key === "Delete") {
            this.delete(this.currentLink, !ctrlKey)
        } else if (key === "Escape") {
            this.dismiss()
        } else if (keyCode === 8) {
            this.setState({ term: this.state.term.slice(0, -1) }, this.getResource)
        } else if (key.length === 1 && !ctrlKey && !altKey && !metaKey) {
            this.setState({ term: this.state.term + key }, this.getResource)
        } else {
            return
        }
        event.preventDefault();
    }

    move = up => {
        let items = document.getElementsByClassName("item")
        let idx = Array.from(items).indexOf(document.activeElement)
        if (idx > -1) {
            up ?
                items[(idx + items.length - 1) % items.length].focus() :
                items[(idx + 1) % items.length].focus()
        }
    }

    render = () => {
        let { resource, term, itemlist, displayDevice} = this.state
        let elements = []
        if (resource) {
            elements.push(
                div(
                    { className: "panel"}, 
                    clock(),
                    itemlist.map(item => { return notifierItem(item)}),
                    battery(displayDevice)
                ),
                React.createElement('hr'),
                resourceHead(resource),
                term && div({className: "searchbox"}, materialIcon("search"), term),
                links(resource, this.activate, this.select)
            )
        }
        return frag( elements)
    }
}

ReactDOM.render(React.createElement(Refude), document.getElementById('app'))