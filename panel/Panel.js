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
import Axios from "axios";
import { ipcRenderer } from 'electron'
import { Clock } from './clock'
import { Battery } from './battery'
import { NotifierItem } from './notifieritems'
import { CloseButton } from "./closebutton";
import { MinimizeButton } from "./minimizebutton";
import { Flash} from "./Flash"
import { Resource, Term, Links} from "./Resource"
import './Refude.css'
import '../common/common.css'
import { linkHref } from '../common/utils';

export default class Panel extends React.Component {
    content = React.createRef()

    constructor(props) {
        super(props)
        this.history = []
        this.state = { term: "", itemlist: []}
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
            this.getItemlist()
            this.getFlash()
        }

        evtSource.onmessage = event => {
            if (this.resourceUrl === "http://localhost:7938" + event.data) {
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
        this.resourceUrl && Axios.get(`${this.resourceUrl}?term=${this.state.term}`)
            .then(({data:res}) => {
                this.setState({ resource: res })
            })
            .catch(error => this.setState({resource: undefined}))
    }

    getDisplayDevice = () => {
        Axios.get("http://localhost:7938/device/DisplayDevice")
            .then(({data: resource}) => {
                this.setState({ displayDevice: resource.data})
            })
            .catch(error => this.setState({ displayDevice: undefined }))
    }

    getItemlist = () => {
        Axios.get("http://localhost:7938/item/list")
            .then(({data:collection}) => {
                this.setState({itemlist: collection.data})
            })
            .catch(error => this.setState({itemlist: []}))
    }

    getFlash = () => {
        Axios.get("http://localhost:7938/notification/list")
            .then(({data: resource}) => {
                let criticalRes = resource.data.find(r => r.data.Urgency == "Critical")
                if (criticalRes) {
                    this.setState({flash: criticalRes})
                } else {
                    let nowInMillis = new Date().getTime()
                    let first = resource.data[0] 
                    if ( first && first.data.Created > nowInMillis - 6000) {
                        this.setState({flash: first})
                        setTimeout(this.getFlash, first.data.Created + 6020 - nowInMillis)
                    } else {
                        this.setState({flash: undefined})
                    }
                }
            })
            .catch(error => this.setState({flash: undefined}))
    }


    select = link => {
        this.currentLink = link
    }
    
    activate = (link, dismiss) => {
        if (link.rel === "org.refude.defaultaction" || link.rel === "org.refude.action") {
            Axios.post(link.href).then(() => dismiss && this.dismiss())
        } else if (link.rel === "org.refude.delete") {
               Axios.delete(path).then(() => dismiss && this.dismiss())
        } else if (link.rel === "related") {
            Axios.get(link.href)
                .then(({data}) => {
                    if (data && data._links) {
                        let dfAct = data._links.find(l => l.rel === "org.refude.defaultaction") 
                        if (dfAct) {
                            Axios.post(dfAct.href).then(() => dismiss && this.dismiss())
                        }
                    }
                })
        }
    }

    delete = (link, dismiss) => {
        if (link.rel === "related") {
            Axios.get(link.href)
                .then(({data}) => {
                    if (data && data._links) {
                        let deleteLink = data._links.find(l => l.rel === "org.refude.delete") 
                        if (deleteLink) {
                            Axios.delete(deleteLink.href).then(() => dismiss && this.dismiss())
                        }
                    }
                })
        }
    }

    dismiss = () => {
        this.resourceUrl = undefined
        this.setState({ resource: undefined, term: "" })
    }

    navigateTo = href => {
        this.history.unshift(this.resourceUrl)
        this.resourceUrl = href
        this.setState({ term: "" }, this.getResource)
    }

    navigateBack = () => {
        this.resourceUrl = this.history.shift() 
        this.setState({ term: ""}, () => this.resourceUrl ? this.getResource() : this.dismiss())
    }

    onKeyDown = (event) => {
        let { key, keyCode, ctrlKey, altKey, metaKey } = event;
        if (key === "ArrowRight" || key === "l" && ctrlKey) {
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
        let { itemlist, displayDevice, flash, resource, term } = this.state
        return <div className="refude" id="content" ref={this.content}>
            <div className="panel">
                <Clock />
                {itemlist.map(res => <NotifierItem key={linkHref(res)} res={res} />)}
                <Battery displayDevice={displayDevice} />
                <MinimizeButton />
                <CloseButton />
            </div>
            { !resource && flash && <Flash flash={flash}/> }
            { resource && <Resource resource={resource} term={term} activate={this.activate} select={this.select}/>}
        </div>
    }
}

ReactDOM.render(<Panel />, document.getElementById('app'))
