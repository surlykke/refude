/*
 * Copyright (c) Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

import {doPost} from './utils.js'
import { setNavigation, onKeyDown } from "./navigation.js"
import { div, materialIcon, hr, input, frag } from "./elements.js"
import { clock } from './clock.js'
import { notifierItem } from './notifieritem.js'
import { battery } from './battery.js'
import {resourceHead} from './resource.js'
import {linkDivs} from './linkdiv.js'

const startUrl = "/search/desktop"

export class Refude extends React.Component {
    
    constructor(props) {
        super(props)
        this.history = []
        this.resourceUrl = startUrl 
        this.state = { term: "", itemlist: []}
        setNavigation({goTo: this.goTo, goBack: this.goBack, dismiss: this.dismiss})
        this.watchSse()
    }

    goTo = href => {
        this.history.unshift(this.resourceUrl)
        this.resourceUrl = href
        this.setState({ term: "" }, this.getResource)
    }

    goBack = () => {
        this.resourceUrl = this.history.shift() || this.resourceUrl
        this.setState({ term: "" }, this.getResource)
    }

    dismiss = () => {
        this.history = []
        this.resourceUrl = startUrl 
        this.setState({ term: ""})
        doPost("http://localhost:7938/client/dismiss")
    }

    componentDidMount = () => {
        document.addEventListener("keydown", onKeyDown)
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
        console.log("Fetching", href)
        fetch(href)
            .then(resp => resp.json())
            .then(
                json => {this.setState({resource: json})},
                error => {this.setState({ resource: undefined })}
            )
    }

    handleInputFocus = e => {
        this.getResource()
    } 

    handleInput = e => {
        this.setState({term: e.target.value}, this.getResource)
    }

    render = () => {
        let { term, resource, itemlist, displayDevice} = this.state
        let elements = []
        if (resource) {
            elements.push(
                div(
                    { className: "panel"}, 
                    clock(),
                    itemlist.map(item => { return notifierItem(item)}),
                    battery(displayDevice)
                ),
                resourceHead(resource),
                input({
                    type: 'text',
                    className:'search-box', 
                    value: term,
                    onFocus: () => this.handleInputFocus,
                    onInput: this.handleInput, 
                    autoFocus: true}),
                linkDivs(resource, this.activate)
            )
        }
        return frag(elements)
    }
}

ReactDOM.render(React.createElement(Refude), document.getElementById('app'))
