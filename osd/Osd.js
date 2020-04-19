// Copyright (c) Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import ReactDOM from 'react-dom'
import { getUrl, monitorSSE } from '../common/monitor';
import { ipcRenderer } from 'electron';
import Axios from 'axios'

export class Osd extends React.Component {
    content = React.createRef()

    constructor(props) {
        super(props);
        this.state = {}
        monitorSSE("osd", this.update, this.update, () => this.setState({ event: undefined }))
    };

    componentDidUpdate = () => {
        let div = document.getElementById("osdDiv")
        if (div) {
            ipcRenderer.send("osdSize", { 
                width: div.getBoundingClientRect().width,
                height: div.getBoundingClientRect().height
            })
        }
    }


    update = () => {
        Axios.get(`http://localhost:7938/osd`)
            .then(resp => {
                console.log("osd set state:", {event: resp.data})
                this.setState({event: resp.data})
            })
            .catch(err => {
                console.error(err)
                this.setState({event: undefined})
            })
    }


    render = () => {
        let { event } = this.state
        if (event) {
            let style = {
                width: "fit-content"
            }
            let messageStyle = {
                overflow: "hidden",
                whiteSpace: "nowrap",
                marginRight: "6px",
            };

            let titleStyle = {
                fontSize: "1em",
            };
            
            let bodyStyle = {
                fontSize: "0.9em",
            };

            let iconStyle = {
                float: "left",
                marginRight: "6px"
            }

            let iconUrl = ""
            if (event.IconName) {
                iconUrl = `http://localhost:7938/icon?name=${event.IconName}&theme=oxygen`
            }

            ipcRenderer.send("osdShow", true)
            console.log("osd render with", event)
            return <div id="osdDiv" style={style}>
                    <img width="24px" height="24px" style={iconStyle} src={iconUrl} alt="" />
                    <div style={messageStyle}>
                        <div style={titleStyle}>{event.Title}</div>
                        {event.Message.map(m => <div style={bodyStyle}>{m}</div>)}
                    </div>
                </div>
        } else {
            ipcRenderer.send("osdShow", false)
            return <div id="osdDiv" ></div>;
        }
    }
}

ReactDOM.render(<Osd />, document.getElementById('osd'))
