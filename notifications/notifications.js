// Copyright (c) Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import { getUrl, monitorSSE } from "../common/monitor";

export class Notifications extends React.Component {
    constructor(props) {
        super(props);
        this.state = { events: [], shown: true}
        monitorSSE("events", this.update, this.update, () => {this.setState({events: []})})
    };

    update = () => {
        getUrl("/notifications", resp => {
            this.setState({ events: resp.data })
        })
    }

    render = () => {

        let nameStyle = {
            overflow: "hidden",
            whiteSpace: "nowrap",
            marginRight: "6px",
        };

        let commentStyle = {
            fontSize: "0.8em",
        };


        let iconStyle =  {
            float: "left",
            marginRight: "6px"
        }

        let iconUrl = (res) => {
            if (res.IconName) {
                return `http://localhost:7938/icon?name=${res.IconName}&theme=oxygen`
            } else {
                return ""
            }
        }
        let items = this.state.events.map(e => { 
            return <div id={e.Self} >
                        <img width="24px" height="24px" style={iconStyle} src={iconUrl(e)} alt="" />
                        <div style={nameStyle}>{e.Title}</div>
                        <div style={commentStyle}>{e.Comment}</div>
                    </div>
        })

        if (this.state.shown) {
            return  <div> {items} </div>
        } else {
            return null;
        }
    }
}
 