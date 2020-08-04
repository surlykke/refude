// Copyright (c) Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import ReactDOM from 'react-dom'
import { monitorPath, iconUrl } from '../common/monitor';
import { ipcRenderer } from 'electron';
import Axios from 'axios'
import './Osd.css'

let acceptNotFound = status => {
    return (status >= 200 && status < 300) || status === 404; // default
}

export class Osd extends React.Component {
    content = React.createRef()

    constructor(props) {
        super(props);
        this.state = {}
        monitorPath("/notification/osd", this.update, this.update, () => this.setState({ event: undefined }))
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
        Axios.get(`http://localhost:7938/notification/osd`)
            .then(resp => {
                console.log("Setting data: ", resp.data)
                this.setState({ event: resp.data })
            })
            .catch(err => {
                console.error(err)
                this.setState({ event: undefined })
            })
    }


    render = () => {
        let { event } = this.state
        if (event) {
            return <div id="osdDiv" className="osd">
                <div id="iconDiv" className="icon">
                    <img width="100%" height="100%" src={iconUrl(event.IconName)} alt="" />
                </div>
                <div id="messageDiv" className="message">
                    {
                        event.Gauge !== undefined ?
                        <meter min="0" max="100" value={event.Gauge}></meter> :
                        <> 
                            <div className="title">
                                {event.Title}
                            </div>
                            {event.Message.map((m, i) =>
                                <div key={`line${i}`} className="body">
                                    {m}
                                </div>)
                            }
                        </>
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
