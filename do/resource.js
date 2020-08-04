//Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.

import React from 'react';
import { keyDownHandler } from './keyhandler';
import './Do.css'
import { iconUrl } from '../common/monitor';

export class Resource extends React.Component {
    constructor(props) {
        super(props)
        this.state = {index: 0}
        this.actLen = props.res.Actions.length
    }

    componentDidMount = () => document.getElementById(this.props.res.Self).focus()

    curAct = () => this.props.res.Actions[this.state.index]

    up = () => this.setState({index: (this.state.index + this.actLen - 1) % this.actLen})
    down = () => this.setState({index: (this.state.index + 1) % this.actLen })
    post = () => this.curAct() && this.props.post(this.curAct().Path) 
    keyHandler = keyDownHandler(this.up, this.down, this.props.back, undefined, this.post, undefined, this.props.dismiss)

    render = () => {
        let { res } = this.props;

        let iconClassName = "icon"
        if (res.Type === "window") {
            iconClassName += " window"
            if (res.Data.States && res.Data.States.includes("_NET_WM_STATE_HIDDEN")) {
                iconClassName += " minimized"
            }
        }
             
        let actionClassName = i => i === this.state.index ? "action selected" : "action"

        return <>
                <div key="body" id={res.Self} className="item" tabIndex="-1" onKeyDown={this.keyHandler}>
                    <img width="24px" height="24px" className={iconClassName} src={iconUrl(res.IconName)} alt="" />
                    <div className="name">{res.Title}</div>
                    <div className="comment">{res.Comment}</div>
                </div>
                <fieldset className="group">
                    <legend>Actions</legend>
                    {res.Actions.map((a, i) => 
                        <div key={a.Path} className={actionClassName(i)}> 
                            <img className={iconClassName} src={iconUrl(a.IconName)} alt="" height="18" width="18"/>
                            <div key={i}> {a.Title}</div> 
                        </div>
                    )}
                </fieldset>
            </>
    }
}

