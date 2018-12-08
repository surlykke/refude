// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';
import {Notification} from "./notification";
import {monitorResources} from "../common/monitor";
import {T} from "../common/translate";

let gui = window.require('nw.gui');

export default class Notifications extends React.Component {
    constructor(props) {
        super(props);
        this.state = {notifications: []}
        monitorResources(
            "notifications-service",
            "application/vnd.org.refude.desktopnotification+json",
            notifications => this.setState({notifications: notifications})
        );
    }

    select = (item) => {
    };

    execute = (item) => {
    };

    dismiss = () => {
        gui.App.quit();
    };

    cancel = () => {
        this.setState({selected: undefined});
    };

    cancelOnEscape = (event) => {
        if (event.key === 'Escape') {
            this.cancel();
        }
    };

    render = () => {
        let divStyle = {
            width: "100%",
            height: "100%",
            textAlign: "center",
            display: "flex",
            alignItems: "center",
            justifyContent: "center"
        };

        if (this.state.notifications.length > 0) {
            return (
                <div>
                    {this.state.notifications.map(n => (<Notification key={n.Id} notification={n}/>))}
                </div>
            );
        } else {
            return <div style={divStyle}><div>{T("No notifications")}</div></div>
        }

    }
}
