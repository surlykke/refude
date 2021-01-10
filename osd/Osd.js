// Copyright (c) Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import ReactDOM from 'react-dom'
import { monitorPath, findLink, path2Url } from '../common/monitor';
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
    };

    componentDidMount = () => {
        ipcRenderer.on("sseopen", this.update)
        ipcRenderer.on("sseerror", this.error)
        ipcRenderer.on("/notification/osd", this.update)
    }

    componentWillUnmount = () => {
        ipcRenderer.removeListener("sseopen", this.update)
        ipcRenderer.removeListener("sseerror", this.error)
        ipcRenderer.removeListener("/notification/osd", this.update)
    }

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
                let self = findLink(resp.data, "self")
                this.setState({ event: resp.data, self: self })
            })
            .catch(err => {
                console.error(err)
                this.error()
            })
    }

    error = () => {
        this.setState({event: undefined})
    }

    render = () => {
        let { event, self } = this.state
        if (event && self) {
            return <div id="osdDiv" className="osd">
                <div id="iconDiv" className="icon">
                    <img width="100%" height="100%" src={path2Url(self.icon)} alt="" />
                </div>
                <div id="messageDiv" className="message">
                    {
                        event.Type === "gauge" ?
                        <meter min="0" max="100" value={event.Gauge}></meter> :
                        <> 
                            <div className="title">
                                {event.Subject}
                            </div>
                            {event.Body.map((m, i) =>
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
