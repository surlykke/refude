//Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.

import React from 'react';
import { keyDownHandler } from './keyhandler';
import { postUrl } from '../common/monitor';
import { ipcRenderer } from 'electron';

export class Resource extends React.Component {
    constructor(props) {
        super(props)
        this.state = {index: 0}
        this.actLen = props.res.Actions.length
    }

    componentDidUpdate = () => {
        console.log("Resource did update");
    }

    componentDidMount = () => {
        console.log("Resource did mount", document.getElementById(this.props.res.Self))
        document.getElementById(this.props.res.Self).focus()
    }

    curAct = () => this.props.res.Actions[this.state.index]

    up = () => this.setState({index: (this.state.index + this.actLen - 1) % this.actLen})
    down = () => this.setState({index: (this.state.index + 1) % this.actLen })
    post = () => this.curAct() && this.props.post(this.curAct().Path) 

    activate = () => {
        let action = this.props.res.Actions[this.state.index]
        action && postUrl(action.Path, response => ipcRenderer.send("doClose"));
    }

    keyHandler = keyDownHandler(this.up, this.down, this.props.back, undefined, this.post, undefined, this.props.dismiss)

    render = () => {
        let { res } = this.props;

        let style = {
            position: "relative",
            marginRight: "5px",
            padding: "4px",
            paddingRight: "18px",
            verticalAlign: "top",
            overflow: "hidden",
            //height: "30px",
            outline: "none",
        };

        let iconStyle = {
            float: "left",
            marginRight: "6px"
        };

        let actionStyle = {
            marginLeft: "25px",
            fontSize: "11px",
            width: "75%",
            padding: "6px",
            paddingRight: "18px",
        }
        
        let selectedActionStyle = Object.assign(
            {
                border: "solid 2px black", 
                borderRadius: "6px",
            },
            actionStyle
        )


        if (res.Type === "window") {
            Object.assign(iconStyle, {
                WebkitFilter: "drop-shadow(5px 5px 3px grey)",
                overflow: "visible"
            });

            if (res.Data.States && res.Data.States.includes("_NET_WM_STATE_HIDDEN")) {
                Object.assign(iconStyle, {
                    marginLeft: "10px",
                    marginTop: "10px",
                    width: "14px",
                    height: "14px",
                    opacity: "0.7"
                })
            }
        }

        let iconUrl = res.IconName ? `http://localhost:7938/icon?name=${res.IconName}&theme=oxygen` : '';

        let nameStyle = {
            overflow: "hidden",
            whiteSpace: "nowrap",
            marginRight: "6px",
        };

        let commentStyle = {
            fontSize: "0.8em",
        };

        return <>
                <div key="body" id={res.Self} style={style} tabIndex="-1" onKeyDown={this.keyHandler}>
                    <img width="24px" height="24px" style={iconStyle} src={iconUrl} alt="" />
                    <div style={nameStyle}>{res.Title}</div>
                    <div style={commentStyle}>{res.Comment}</div>
                </div>
                {res.Actions.map((a, i) => {
                        let s = i === this.state.index ? selectedActionStyle : actionStyle
                        return <div key={a.Path}> 
                            <div key={i} style={s}>{a.Title}</div> 
                        </div>
                    }
                )}
            </>
    }
}

