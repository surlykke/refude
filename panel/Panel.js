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
import { Flash} from "./Flash"
import { Resource, Term, Links} from "./Resource"
import { getUrl, addParam, postUrl, deleteUrl } from '../common/monitor'
import './Refude.css'
import '../common/common.css'

export default class Panel extends React.Component {
    content = React.createRef()

    constructor(props) {
        super(props)
        this.history = []
        this.state = { term: "", itemLinks: [] }
        ipcRenderer.on("show", (evt, up) => {
            if (!this.resourceUrl) {
                this.setState({ term: "" })
                this.resourceUrl = "http://localhost:7938/search/desktop"
                this.getResource()
            } else if (document.hasFocus()) {
                this.move(up)
            }
        })
        this.watchSse()
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
            this.getDisplayDevice()
            this.getItemLinks()
            this.getFlash()
        }

        evtSource.onmessage = event => {
            console.log("sse:", event)
            if (this.resourceUrl === "http://localhost:7938" + event.data) {
                this.getResource()
            } else if ("/device/DisplayDevice" === event.data) {
                this.getDisplayDevice()
            } else if ("/item/list" === event.data) {
                this.getItemLinks()
            } else if ("/notification/flash" === event.data) {
                this.getFlash()
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

        document.addEventListener("keydown", this.onKeyDown)
     };

    getResource = () => {
        console.log("resource:", this.resourceUrl)
        let resourceUrl = this.resourceUrl
        if (resourceUrl) {
            console.log("Getting ", resourceUrl)
            getUrl(addParam(resourceUrl, "term", this.state.term), ({ data }) => {
                if (resourceUrl === this.resourceUrl) {  // may have changed while request in flight
                    this.setState({ resource: data })
                }
            })
        }
    }

    getDisplayDevice = () => {
        console.log("getDisplayDevice")
        getUrl("http://localhost:7938/device/DisplayDevice",
            resp => this.setState({ displayDevice: resp.data }),
            () => this.setState({ displayDevice: undefined }))
    }

    getItemLinks = () => {
        getUrl("http://localhost:7938/item/list",
            ({data}) => this.setState({itemLinks: data._links.filter(l => l.rel == "related")}),
            () => this.setState({ itemLinks: [] })
        )
    }

    getFlash = () => {
        getUrl(
            "/notification/flash",
            ({data}) => this.setState({flash: data}),
            () => this.setState({flash: undefined})
        )
    }


    select = link => {
        this.currentLink = link
    }
    
    activate = (link, dismiss) => {
        let dismissIf = () => dismiss && this.dismiss()
        if (link.rel === "org.refude.defaultaction" || link.rel === "org.refude.action") {
            postUrl(link.href, dismissIf)
        } else if (link.rel === "org.refude.delete") {
            deleteUrl(link.href, dismissIf)
        } else if (link.rel === "related") {
            getUrl(link.href, ({data}) => {
                if (data && data._links) {
                    let dfAct = data._links.find(l => l.rel === "org.refude.defaultaction") 
                    if (dfAct) {
                        postUrl(dfAct.href, dismissIf)
                    }
                }
            })
        }
        console.log(link)
    }

    delete = (link, dismiss) => {
        console.log("delete, link:", link)
        let doAfter = () => dismiss && this.dismiss()
        if (link.rel === "related") {
            getUrl(link.href, ({data}) => {
                if (data && data._links) {
                    console.log("Got resource", data)
                    let deleteLink = data._links.find(l => l.rel === "org.refude.delete") 
                    console.log("deleteLink:", deleteLink)
                    if (deleteLink) {
                        deleteUrl(deleteLink.href, doAfter)
                    }
                }
            })
        }
    }

    dismiss = () => {
        console.log("dismissing...")
        this.resourceUrl = undefined
        this.setState({ resource: undefined, term: "" })
    }

    navigateTo = href => {
        this.history.unshift([this.resourceUrl, this.state.term])
        this.resourceUrl = href
        this.setState({ term: "" }, this.getResource)
    }

    navigateBack = () => {
        let term
        [this.resourceUrl, term] = this.history.shift() || []
        console.log("back to:", this.resourceUrl)
        this.setState({ term: term }, () => this.resourceUrl ? this.getResource() : this.dismiss())
    }

    onKeyDown = (event) => {
        let { key, keyCode, ctrlKey, altKey, shiftKey, metaKey } = event;
        /*if (key === "i" && ctrlKey && shiftKey) {
            ipcRenderer.send("devtools")
        } else*/ if (key === "ArrowRight" || key === "l" && ctrlKey) {
            console.log("currentLink:", this.currentLink)
            this.currentLink.rel === "related" && this.navigateTo(this.currentLink.href)
        } else if(key === "ArrowLeft" || key === "h" && ctrlKey) { 
            this.navigateBack()
        } else if (key === "ArrowUp" || key === "k" && ctrlKey) {
            this.move(true)
        } else if (key === "ArrowDown" || key === "j" && ctrlKey) {
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
        let { itemLinks, displayDevice, flash, resource, term } = this.state
        console.log("panel render, resource:", resource)
        return <div className="refude" id="content" ref={this.content}>
            <div className="panel">
                <Clock />
                {itemLinks.map(itemLink => <NotifierItem key={itemLink.href} itemLink={itemLink} />)}
                <Battery displayDevice={displayDevice} />
                <MinimizeButton />
                <CloseButton />
            </div>
            { !resource && flash && <Flash flash={flash}/> }
            { resource && <>
                <Resource resource={resource}/>
                <Term term={term}/>
                <Links links={resource._links.filter(l => "self" !== l.rel)} activate={this.activate} select={this.select}/>
              </>
            }
        </div>
    }
}

ReactDOM.render(<Panel />, document.getElementById('app'))
