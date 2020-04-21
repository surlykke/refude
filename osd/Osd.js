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
            ipcRenderer.send("osdShow", {
                width: div.getBoundingClientRect().width,
                height: div.getBoundingClientRect().height
            })
        }
    }


    update = () => {
        Axios.get(`http://localhost:7938/osd`)
            .then(resp => {
                this.setState({ event: resp.data })
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
                padding: "0.5em",
                width: "22em",
                display: "flex",
                backgroundColor: "white"
            }
            let iconStyle = {
                width: "3em",
                height: "3em",
                padding: "0px",
                margin: "0px"

            }
   
            let messageStyle = {
                marginLeft: "0.7em",
                overflow: "hidden",
                marginRight: "6px",
                flex: "1",
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

            return <div id="osdDiv" style={style}>
                <div style={iconStyle}>
                    <img width="100%" height="100%" src={iconUrl} alt="" />
                </div>
                <div style={messageStyle}>
                    <div style={titleStyle}>
                        {event.Title}
                    </div>
                    {event.Message.map(m =>
                        <div style={bodyStyle}>
                            {m}
                        </div>)
                    }
                </div>
            </div>
        } else {
            ipcRenderer.send("osdHide")
            return null;
        }
    }
}

ReactDOM.render(<Osd />, document.getElementById('osd'))
