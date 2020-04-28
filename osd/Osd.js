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


let acceptNotFound = status => {
    return (status >= 200 && status < 300) || status === 404; // default
}

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
            let [height, width] = [0, 0]
            Array.from(div.children).forEach(c => {
                height = Math.max(height, c.getBoundingClientRect().height)
                width = width + c.getBoundingClientRect().width; 
            })
            ipcRenderer.send("osdShow", {width: width, height: height})
        }
    }


    update = () => {
        Axios.get(`http://localhost:7938/notification/osd`, {validateStatus: acceptNotFound})
            .then(resp => {
                if (resp.status === 404) {
                    this.setState({event: undefined})
                } else {
                    this.setState({ event: resp.data })
                }
                
            })
            .catch(err => {
                console.error(err)
                this.setState({ event: undefined })
            })
    }


    render = () => {
        let { event } = this.state
        if (event) {
            let style = {
                position: "relative",
                backgroundColor: "white",
                width: "30em",
                height: "7em",
            }
            let iconStyle = {
                position: "absolute",
                width: "2.2em",
                height: "2.2em",
                padding: "0.5em",
            }

            let messageStyle = {
                position: "absolute",
                left: "3.2em",
                overflow: "hidden",
                padding: "0.5em 0.5em 0.5em 0em",
                minWidth: "6em",
                maxWidth: "18em",
            };

            let titleStyle = {
                fontSize: "0.9em",
                whiteSpace: "nowrap",
            };

            let bodyStyle = {
                maxHeight: "2.2em",
                fontSize: "0.7em",
                paddingTop: "0.2em",
                paddingBottom: "0.2em",
                overflowY: "hidden",
            };

            let iconUrl = event.IconName && `http://localhost:7938/icon?name=${event.IconName}&theme=oxygen&size=48`

            let messageDiv
            if (event.Gauge !== undefined) {
                messageDiv = 
                    <div id="messageDiv" style={messageStyle}>
                        <meter min="0" max="100" value={event.Gauge}></meter>
                    </div>
            } else {
                messageDiv =
                   <div id="messageDiv" style={messageStyle}>
                        <div style={titleStyle}>
                            {event.Title}
                        </div>
                        {event.Message.map((m, i) =>
                            <div key={`line${i}`} style={bodyStyle}>
                                {m}
                            </div>)
                        }
                    </div>
            }

            return <div id="osdDiv" style={style}>
                <div id="iconDiv" style={iconStyle}>
                    <img width="100%" height="100%" src={iconUrl} alt="" />
                </div>
                {messageDiv} 
            </div>
        } else {
            ipcRenderer.send("osdHide")
            return null;
        }
    }
}

ReactDOM.render(<Osd />, document.getElementById('osd'))
