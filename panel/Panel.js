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
import { Flash} from "../flash/Flash"
import { Resource, move } from "../resource/Resource"
import { getUrl, addParam } from '../common/monitor'
import { makeResourceController } from '../resource/ResourceController'
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
                this.resourceUrl = "/search/desktop"
                this.getResource()
            } else if (document.hasFocus()) {
                move(up)
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
            if (this.resourceUrl === event.data) {
                this.getResource()
            } else if ("/device/DisplayDevice" === event.data) {
                this.getDisplayDevice()
            } else if ("/items" === event.data) {
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

        document.addEventListener("keydown", this.resourceController.onKeyDown)
    };

    getResource = () => {
        let resourceUrl = this.resourceUrl
        if (resourceUrl) {
            getUrl(addParam(resourceUrl, "term", this.state.term), ({ data }) => {
                if (resourceUrl === this.resourceUrl) {  // may have change while request in flight
                    this.setState({ resource: data })
                }
            })
        }
    }

    getDisplayDevice = () => {
        console.log("getDisplayDevice")
        getUrl("/device/DisplayDevice",
            resp => this.setState({ displayDevice: resp.data }),
            () => this.setState({ displayDevice: undefined }))
    }

    getItemLinks = () => {
        getUrl("/items",
            ({data}) => this.setState({itemLinks: data._related}),
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


    dismiss = () => {
        this.resourceUrl = undefined
        this.setState({ resource: undefined, term: "" })
    }

    navigateTo = link => {
        this.history.unshift([this.resourceUrl, this.state.term])
        this.resourceUrl = link.href
        this.setState({ term: "" }, this.getResource)
    }

    navigateBack = () => {
        let term
        [this.resourceUrl, term] = this.history.shift() || []
        this.setState({ term: term }, () => this.resourceUrl ? this.getResource() : this.dismiss())
    }

    handleKey = key => {
        if (key === "Backspace") {
            this.setState({ term: this.state.term.slice(0, -1) }, this.getResource)
        } else if (key.length === 1) {
            this.setState({ term: this.state.term + key }, this.getResource)
        }
    }

    resourceController = makeResourceController(this.navigateTo, this.navigateBack, this.dismiss, this.handleKey)

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
            <Flash flash={!resource && flash}/> 
            <Resource resource={resource} controller={this.resourceController} term={term} />
        </div>
    }
}

ReactDOM.render(<Panel />, document.getElementById('app'))
