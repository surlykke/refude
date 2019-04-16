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
import {WIN, watchScreenChanges, /*devtools,*/ loadPosition, subscribe} from '../common/utils'
import {Clock} from './clock'
import {Battery} from './battery'
import {NotifierItems} from './notifieritems'
import {Notificationalert} from "./notificationalert";
import {DragField} from './dragfield'
import {CloseButton} from "./closebutton";
import {Do} from './do'

const Window = nw.Window.get();

const style = {
    margin: "0px",
    paddingLeft: "6px",
    paddingRight: "6px",
    paddingTop: "2px",
    paddingBottom: "2px",
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
        //devtools();
        super(props)
        WIN.on('loaded', () => {
            loadPosition();
            watchScreenChanges();
        });
    }

    componentDidMount = () => {
        subscribe("componentUpdated", this.adjustSize);
        this.adjustSize();
    };


    adjustSize = () => {
        setTimeout(
            () => {
                let {width, height} = this.content.getBoundingClientRect()
                Window.resizeTo(Math.round(width) - 1, Math.round(height))
            },
            10
        )
    };

    render = () =>
        <div style={{width: "500px"}}>
            <div style={style} id="content" ref={div => { this.content = div }}>
                <Clock style={pluginStyle}/>
                <Notificationalert style={pluginStyle}/>
                <NotifierItems style={pluginStyle}/>
                <Battery style={pluginStyle}/>
                <DragField style={pluginStyle}/>
                <CloseButton style={pluginStyle}/>
                <Do/>
            </div>
        </div>
}

