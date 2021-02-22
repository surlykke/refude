// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
/*
 * Copyright (c) Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
import React from 'react'
import ReactDOM from 'react-dom'
import { ipcRenderer } from 'electron'
import { Clock } from './clock'
import { Battery } from './battery'
import { NotifierItem } from './notifieritems'
import { CloseButton } from "./closebutton";
import { MinimizeButton } from "./minimizebutton";
import { Osd } from "../osd/Osd"
import { GenericResource } from "../resource/GenericResource"
import { move, LinkList } from "../resource/LinkList"
import './Refude.css'
import '../common/common.css'
import { getUrl, postUrl, deleteUrl, addParam, findLink, splitLinks } from '../common/monitor'
import { link } from 'fs'
import { throwDeprecation } from 'process'

export default class Panel extends React.Component {
    content = React.createRef()

    constructor(props) {
        super(props)
        this.appState = {} // What state we have that is not relevant for React's rendering
        this.history = [] // appState 'stack'
        this.state = { term: "", itemLinks: [] }
        ipcRenderer.on("show", (evt, up) => {
            if (!this.appState.resourceUrl) {
                this.setState({ term: "" })
                this.appState.resourceUrl = "/search/desktop"
                this.getResource()
            } else {
                move(up)
            }
        })
        this.watchSse()
        this.getDisplayDevice()
        this.getItemLinks()
        this.getOsd()
    }

    watchSse = () => {
        let evtSource = new EventSource("http://localhost:7938/watch")

        evtSource.onerror = event => {
            ipcRenderer.send("sseerror")
            if (evtSource.readyState === 2) {
                setTimeout(watchSse, 5000)
            }
        }

        evtSource.onopen = () => {
        }

        evtSource.onmessage = event => {
            if (this.appState.resourceUrl === event.data) {
                this.getResource()
            } else if ("/device/DisplayDevice" === event.data) {
                this.getDisplayDevice()
            } else if ("/items" === event.data) {
                this.getItemLinks()
            } else if ("/notification/osd" === event.data) {
                this.getOsd()
            }
        }
    }

    componentDidMount = () => {
        new ResizeObserver((observed) => {
            let content = document.getElementById('content');
            if (content) { // shouldn't be nessecary 
                let width =
                    parseFloat(window.getComputedStyle(content).getPropertyValue("width")) +
                    parseFloat(window.getComputedStyle(content).getPropertyValue("padding-left")) +
                    parseFloat(window.getComputedStyle(content).getPropertyValue("padding-right"))
                let height =
                    parseFloat(window.getComputedStyle(content).getPropertyValue("height")) +
                    parseFloat(window.getComputedStyle(content).getPropertyValue("padding-top")) +
                    parseFloat(window.getComputedStyle(content).getPropertyValue("padding-bottom"))

                ipcRenderer.send("panelSizeChange", width, height);
            }
        }).observe(this.content.current)

        document.addEventListener("keydown", this.keyDownHandler)
    };

    getResource = () => {
        let resourceUrl = this.appState.resourceUrl
        if (resourceUrl) {
            console.log("getting", resourceUrl)
            getUrl(addParam(resourceUrl, "term", this.state.term), ({ data }) => {
                if (resourceUrl === this.appState.resourceUrl) {  // may have change while request in flight
                    let { self, links } = splitLinks(data)
                    this.setState({ resource: data, self: self, links: links })
                }
            })
        }
    }

    getDisplayDevice = () => {
        getUrl("/device/DisplayDevice",
            resp => this.setState({ displayDevice: resp.data }),
            () => this.setState({ displayDevice: undefined }))
    }

    getOsd = () => {
        getUrl("/notification/osd", response => this.setState({ osd: response.data }), err => { console.log("getOsd recieved error:", err); this.setState({ osd: undefined }) })
    }

    getItemLinks = () => {
        console.log("Getting items")
        getUrl("/items", response => this.setState({ itemLinks: response.data }), () => this.setState({ itemLinks: [] }))
    }
    
    highlight = link => {
        console.log("highlighting:", link)
        console.log("link?.profile:", link?.profile, "compare",  "/profile/window" !== link?.profile)
        if ( "/profile/window" !== link?.profile) {
            link = undefined
        }
        if (link?.href !== this.appState.highlightedLink?.href) {
            if (link?.href) {
                console.log("Posting", addParam(link.href, "action", "highlight"))
                postUrl(addParam(link.href, "action", "highlight"))
            } else {
                postUrl("/window/unhighlight")
            }
        }
        this.appState.highlightedLink = link
    }

    selectLink = link => {
        this.appState.focusedLink = link
        link.rel === "related" || (link = this.state.self)
        this.highlight(link)
    }

    clearState = then => {
        console.log("clearState")
        this.appState = {}
        this.setState({ resource: undefined, links: undefined, term: "", self: undefined }, then && then())
    }

    execute = (action, keep) => {
        console.log("execute, action:", !!action, "keep:", keep, "focusedLink:", this.appState.focusedLink)
        if (this.appState.focusedLink) {
            this.highlight(undefined)
            action(this.appState.focusedLink.href, () => keep || this.clearState())
        }
    }

    act = keep => this.execute(postUrl, keep)
    del = keep => this.execute(deleteUrl, keep)
    
    dismiss = () => {
        this.clearState(this.highlight)
        postUrl("/window/unhighlight")
    }
    
    navigateTo = () => {
        console.log("navigateto, rel:", this.appState?.focusedLink?.rel)
        if (this.appState?.focusedLink?.rel === "related") {
            this.history.unshift(this.appState)
            this.appState = {resourceUrl: this.appState.focusedLink.href, term: ""}
            this.getResource()
        }
    }

    navigateBack = () => {
        this.appState = this.history.shift() || {} 
        if (this.appState.resourceUrl) {
            this.getResource()
        } else {
            this.clearState()
        }
    }

    keyDownHandler = (event) => {
        let { key, keyCode, charCode, code, ctrlKey, altKey, shiftKey, metaKey } = event;
        if (key === "i" && ctrlKey && shiftKey) {
            ipcRenderer.send("devtools")
        } else if (key === "ArrowRight" || key === "l" && ctrlKey) {
            this.navigateTo()
        } else if (key === "ArrowLeft" || key === "f" && ctrlKey) {
            this.navigateBack()
        } else if (key === "ArrowUp" || key === "k" && ctrlKey) {
            move(true)
        } else if (key === "ArrowDown" || key === "j" && ctrlKey) {
            move(false)
        } else if (key === "Enter") {
            this.act(ctrlKey)
        } else if (key === "Delete") {
            this.del(ctrlKey)
        } else if (key === "Escape") {
            this.dismiss()
        } else if (keyCode === 8 && !ctrlKey && !altKey && !metaKey) {
            console.log("Setting term to", this.state.term.slice(0, -1))
            this.setState({ term: this.state.term.slice(0, -1) }, this.getResource)
        } else if (key.length === 1 && !ctrlKey && !altKey && !metaKey) {
            console.log("Setting term to", this.state.term + key)
            this.setState({ term: this.state.term + key }, this.getResource)
        } else {
            return
        }
        event.preventDefault();
    }

    render = () => {
        return <div className="refude" id="content" ref={this.content}>
            <div className="panel">
                <Clock />
                {
                    this.state.itemLinks.map(itemLink => <NotifierItem key={itemLink.href} itemLink={itemLink} />)
                }
                <Battery displayDevice={this.state.displayDevice} />
                <MinimizeButton />
                <CloseButton />
            </div>
            {!this.state.resource && this.state.osd && <Osd event={this.state.osd} />}
            {this.state.resource && this.state.self && <GenericResource self={this.state.self} />}
            {this.state.resource && !this.state.self && <div className="searchbox">{this.state.term}</div>}
            {this.state.links &&
                <LinkList links={this.state.links}
                    focusedHref={this.appState.focusedLink?.href}
                    selectLink={this.selectLink}
                    act={this.act} />
            }
        </div>
    }
}


ReactDOM.render(<Panel />, document.getElementById('app'))
