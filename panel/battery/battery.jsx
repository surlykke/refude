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
import {doGetIfNoneMatch} from '../../common/http'

const errorData = {
    style: {
        color: 'red',
        fontWeight: 'normal',
        marginRight: '0.4em'
    },
    percentage: '?'
}

export class Battery extends React.Component {

    constructor(props) {
        super(props);
        this.onUpdated = props.onUpdated;
        this.state = {pct: -1, state: "Unknown"};
        this.etag = undefined;
    }

    componentDidMount = () => {
        let update = () => {
            doGetIfNoneMatch("power-service", "/devices/DisplayDevice", this.etag).then(resp => {
                this.etag = resp.headers.etag
                this.setState({
                    pct: resp.json.State === "Unknown" ? -1 : resp.json.Percentage,
                    state: resp.json.State
                });
            }).catch(err => {
                if (err.status !== 304) {
                    console.log("err:", err);
                    this.setState({pct: 0, state: "Unknown"})
                }
            }).then(setTimeout(update, 1000));
        };
        update();
    };

    componentDidUpdate() {
        this.onUpdated();
    }


    render = () => {
        let {pct, state} = this.state;

        console.log("pct: ", pct, ", state:", state);
        let archEndX = 50 * Math.cos((0.5 + pct * 0.02) * Math.PI);
        let archEndY = -50 * Math.sin((0.5 + pct * 0.02) * Math.PI);
        let bigArchFlag = pct >= 50 ? 1 : 0;
        console.log("archEndX, archEndY:", archEndX, archEndY);
        let color = state === "Charging" || state === "Fully charged" ? "grey" : "lightgrey";
        let dotFill = state === "Fully charged" ? "black" : "white";

        let alert = state === "Unknown" || pct < 5;
        let alertText = state === "Unknown" ? '?' : '!';
        let title = state + " " + Math.round(pct) + "%";

        return <div title={title} style={this.props.style}>
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="-55 -55 110 110">
                <path d={`M 0 0 v -50 A 50 50 0 ${bigArchFlag} 0 ${archEndX} ${archEndY} L 0 0`} fill={color} stroke={color} strokeWidth="4"/>
                <circle cx="0" cy="0" r="20" stroke="none" fill={dotFill}/>
                {alert && <circle cx="0" cy="0" r="50" stroke="red" strokeWidth="10" fill={dotFill}/>}
                {alert && <text textAnchor="middle" x="0" y="30" style={{fontWeight: "bold", fontSize: "90px", fill: "red"}}>{alertText}</text>}
            </svg>
        </div>
    }
}


