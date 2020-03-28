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
import { devtools, manageZoom, subscribe, managePosition } from '../common/utils'
import { Clock } from './clock'
import { Battery } from './battery'
import { NotifierItems } from './notifieritems'
import { DragField } from './dragfield'
import { CloseButton } from "./closebutton";
import { Do } from './do'
import { Notifications } from './notifications'

const http = require('http');

const Window = nw.Window.get();

const style = {
    margin: "0px",
    paddingLeft: "6px",
    paddingRight: "6px",
    width: "fit-content",
    backgroundColor: "rgba(255,255,255,0.8)"
}

const pluginStyle = {
    display: "inline-block",
    height: "100%",
    marginTop: "0px",
    marginLeft: "0px",
    marginBottom: "0px",
    marginRight: "5px",
    verticalAlign: "middle"
}

export default class Panel extends React.Component {

    constructor(props) {
        super(props)
        managePosition()
        manageZoom()
        //devtools();
    }

    componentDidMount = () => {
        subscribe("componentUpdated", this.adjustSize);
        this.adjustSize()
    };

    adjustSize = () => {
        setTimeout(
            () => {
                let { width, height } = this.content.getBoundingClientRect()
                let zoomLevel = document.body.style.zoom || 1
                let newWidth = Math.round(zoomLevel*width) - 1;
                let newHeight = Math.round(zoomLevel*height);
                if (Math.abs(newWidth - Window.width) > 3 || Math.abs(newHeight - Window.height) > 3) {
                    Window.resizeTo(newWidth, newHeight);
                }
            },
            1
        )
    };

    render = () => {
        

        return <div style={{ width: "500px" }}>
            <div style={style} id="content" ref={div => { this.content = div }}>
                <Clock style={pluginStyle} />
                <NotifierItems style={pluginStyle} />
                <Battery style={pluginStyle} />
                <DragField style={pluginStyle} />
                <CloseButton style={pluginStyle} />
                <Notifications/>
                <Do/> 
            </div>
        </div>
    }
}

 
