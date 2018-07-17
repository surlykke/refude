// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
import React from 'react'
import {render} from 'react-dom'
import {doGetH, nwSetup} from '../common/utils'
import {Clock} from './clock/clock'
import {Battery} from './battery/battery'
import {NotifierItems} from './notifieritems/notifieritems'
import {HideButton} from './hidebutton/hidebutton'
import {DragField} from './dragfield/dragfield'
import {Notifications} from './notifications/notifications'

const Window = nw.Window.get();
const style = {
    display: "inline-block",
    margin: "0px",
    padding: "2px",
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

class Panel extends React.Component {

    constructor(props) {
        super(props)

        nwSetup((argv) => {
        })
    }

    onMove = (x, y) => {
        if (this.dispEtag) {
            localStorage.setItem(this.dispEtag + ".x", x);
            localStorage.setItem(this.dispEtag + ".y", y);
        }
    };


    maintainPos = () => {
        doGetH({service: "wm-service", path: "/display", ifNoneMatch: this.dispEtag}, (json, headers) => {
            if (headers && headers.etag) {
                let x = localStorage.getItem(headers.etag + ".x");
                let y = localStorage.getItem(headers.etag + ".y");
                if (x && y) {
                    Window.moveTo(parseInt(x), parseInt(y));
                }
                this.dispEtag = headers.etag;
            }
        });


        setTimeout(this.maintainPos, 5000);
    };


    componentDidMount = () => {
        this.adjustSize();
        Window.on('move', this.onMove);
        setTimeout(this.maintainPos, 10000);
    };


    adjustSize = () => {
        setTimeout(
            () => {
                let {width, height} = this.content.getBoundingClientRect()
                Window.resizeTo(Math.round(width) - 1, Math.round(height))
            },
            10
        )
    }

    render = () =>
        <div style={{height: "100%", width: "500px"}}>
            <div style={style} id="content" ref={div => {
                this.content = div
            }}>
                <Clock style={pluginStyle}/>
                <Battery style={pluginStyle} onUpdated={this.adjustSize}/>
                <NotifierItems style={pluginStyle} onUpdated={this.adjustSize}/>
                <HideButton style={pluginStyle}/>
                <DragField style={pluginStyle}/>
                <Notifications style={pluginStyle} onUpdated={this.adjustSize}/>
            </div>
        </div>
}


render(
    <Panel/>,
    document.getElementById('root')
);


