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

export let Osd = ({ event }) => {
    let self = findLink(event, "self")
    if (self) {
        return <div id="osdDiv" className="osd">
            <div id="iconDiv" className="osd-icon">
                <img width="100%" height="100%" src={path2Url(self.icon)} alt="" />
            </div>
            <div id="messageDiv" className="osd-message">
                {
                    event.Type === "gauge" ?
                        <meter min="0" max="100" value={event.Gauge}></meter> :
                        <>
                            <div className="osd-title">
                                {event.Subject}
                            </div>
                            {event.Body.map((m, i) =>
                                <div key={`line${i}`} className="osd-body">
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
