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
import { DragField } from './dragfield'
import { CloseButton } from "./closebutton";
import { Notifications } from './notifications'
import { ipcRenderer, webFrame} from 'electron'

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
    content = React.createRef()
    
    constructor(props) {
        super(props)
    }

    componentDidMount = () => {
        this.resizeObserver = new ResizeObserver((observed) => {
            if (observed[0] && observed[0].contentRect) {
                let {width, height} = observed[0].contentRect
                let zoom = webFrame.getZoomFactor()
                let data = {width:Math.round(zoom*width), height: Math.round(zoom*height)} 
                ipcRenderer.send("panelSizeChange", data)
            }
        })
        
        this.resizeObserver.observe(this.content.current)
    };

    render = () => {
        return <div style={{ width: "500px" }}>
            <div style={style} id="content" ref={this.content}>
                <Clock style={pluginStyle} />
                <NotifierItems style={pluginStyle} />
                <Battery style={pluginStyle} />
                <DragField style={pluginStyle} />
                <CloseButton style={pluginStyle} />
                <Notifications/>
           </div>
        </div>
    }
}

 
ReactDOM.render(<Panel/>,document.getElementById('app'))
