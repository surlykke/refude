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
import { Clock } from './clock'
import { Battery } from './battery'
import { NotifierItems } from './notifieritems'
import { CloseButton } from "./closebutton";
import { ipcRenderer } from 'electron'
import './Panel.css'
import '../common/common.css'

// Annoyingly, browsers and hence electron, only allows a very limited set of server-sent-event-streams
// so we make do with one and do client-side dispatching.
// We cannot open a sse in the main process (has to be in a render process), so we do it here in panel 
// (which is the main window), forward events to the main process, which, in turn, dispatches to relevant 
// windows. (sigh).
let watchSse = () => {
    console.log("monitorPaths")
    let evtSource = new EventSource("http://localhost:7938/watch")

    evtSource.onerror = event => {
        ipcRenderer.send("sseerror")
        if (evtSource.readyState === 2) {
            setTimeout(watchSse, 5000)
        }
    }

    evtSource.onopen = () => {
        ipcRenderer.send("sseopen")
    }

    evtSource.onmessage = event => {
        ipcRenderer.send("ssemessage", event.data)
    }
}

watchSse()


export default class Panel extends React.Component {
    content = React.createRef()

    constructor(props) {
        super(props)
    }

    componentDidMount = () => {
        this.resizeObserver = new ResizeObserver((observed) => {
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
        })

        this.resizeObserver.observe(this.content.current)
    };

    render = () => {
        return <div style={{ width: "500px" }}>
            <div id="content" className="panel topbar" ref={this.content}>
                <Clock />
                <NotifierItems />
                <Battery />
                <CloseButton />
            </div>
        </div>
    }
}


ReactDOM.render(<Panel />, document.getElementById('app'))
