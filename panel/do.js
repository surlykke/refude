//Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.


import React from 'react'
import { WIN, applicationRank, publish, subscribe } from "../common/utils";
import { ItemList } from "../common/itemlist"
import { Item } from "../common/item"
import { Indicator } from "./indicator";
import { T } from "../common/translate";
import { monitorUrl, getUrl, postUrl } from '../common/monitor';

const http = require('http');

let windowIconStyle = w => {
    let style = {
        WebkitFilter: "drop-shadow(5px 5px 3px grey)",
        overflow: "visible"
    };
    if (w.States.includes("_NET_WM_STATE_HIDDEN")) {
        Object.assign(style, {
            marginLeft: "10px",
            marginTop: "10px",
            width: "14px",
            height: "14px",
            opacity: "0.7"
        })
    }

    return style
};

let timeToExpiry = notification => {
    return Date.parse(notification.Created) - new Date().getTime() + 20000
}

let timeToNextExpiry = notifications => {
    let t
    notifications.forEach(n => {
        if (!(t < timeToExpiry(n))) {
            t = timeToExpiry(n)
        }

    })
    return t
}

class Do extends React.Component {
    constructor(props) {
        //devtools();
        super(props);
        this.state = { notifications: [], windows: [], applications: [], session: { _actions: {} } };

        this.listenForUpDown();
        this.handleBlurEvents();
    };

    componentDidMount = () => {
        monitorUrl("/notifications", resp => {
            if (this.state.notifications.length === 0) { 
                this.notificationFilterScheduler(resp.data)
            } else { // scheduler will already be running
                this.setState({notifications: resp.data.filter(n => timeToExpiry(n) > 0)})
            }
        });
    };

    notificationFilterScheduler = notifications => {
        let filteredNotificatons = notifications.filter(n => timeToExpiry(n) > 0)
        this.setState({ notifications: filteredNotificatons })
        if (filteredNotificatons.length > 0) {
            let delay = timeToNextExpiry(filteredNotificatons) 
            setTimeout(() => this.notificationFilterScheduler(this.state.notifications), delay + 1000)
        }
    }

    componentDidUpdate = () => {
        publish("componentUpdated");
    };

    listenForUpDown = () => {
        let that = this;
        http.createServer(function (req, res) {
            that.showWin();
            if (req.url === "/up") {
                publish("moveRequested", false);
            } else {
                publish("moveRequested", true);
            }
            res.end('')
        }).listen("/run/user/1000/org.refude.panel.do");
    };

    handleBlurEvents = () => {
        WIN.on('focus', () => this.hasfocus = true);
        WIN.on('blur', () => {
            this.hasfocus = false;
            // TAB momentarily unfocuses window - so we wait a bit to see if it's for real
            setTimeout(() => {
                if (!this.hasfocus) publish("dismiss");
            }, 100);
        });
    };

    showWin = () => {
        if (!this.state["shown"]) {
            getUrl("/windows", resp => this.setState({ windows: resp.data.filter(w => w.States.indexOf("_NET_WM_STATE_ABOVE") < 0) }))
            getUrl("/applications", resp => this.setState({ applications: resp.data.filter(app => !app.NoDisplay) }))
            getUrl("/session", resp => this.setState({ session: resp.data }))
            this.setState({ shown: true });
            publish("selectorShown", new Date().getTime())
        }
        WIN.focus();
    };

    select = (url) => {
        publish("windowSelected", this.state.windows.find(w => w._self === url))
    }

    execute = (item) => {
        postUrl(item.url, response => {
            this.reset()
        })
    };

    dismiss = () => {
        if (this.state.shown) {
            // Return focus to whatever was at the top of the (normal) stack
            this.state.windows[0] && postUrl(this.state.windows[0]._self)
        }
        this.reset()
    };

    reset = () => {
        this.setState({ "shown": undefined })
        publish("reset")
    }

    render = () => {
        let itemListStyle = { maxWidth: "300px", maxHeight: "300px" };
        let items = this.state.notifications.map((n, i) => {
            return {
                group: T("Notifications"),
                url: n._self,
                name: n.Subject,
                comment: n.Body,
                image: "http://localhost:7938/" + n.Image,
                showInitially: true,
                rank: 200000 - i*1000
            }
        })

        this.state.windows.forEach((w, i) => items.push({
            group: T("Open windows"),
            url: w._self,
            name: w.Name,
            comment: "",
            image: 'http://localhost:7938/icon/' + w.IconName + "/img",
            iconStyle: windowIconStyle(w),
            showInitially: true,
            rank: 100000 - i*1000
        }))
        this.state.applications.forEach(a => items.push({
            group: T("Applications"),
            url: a._self,
            name: a.Name,
            comment: a.Comment || '',
            image: 'http://localhost:7938/icon/' + a.IconName + "/img",
            rank: 0
        }))
        Object.keys(this.state.session._actions).forEach(key => items.push({
            group: T("Leave"),
            url: this.state.session._self + "?action=" + key,
            name: key,
            comment: this.state.session._actions[key].Description,
            image: 'http://localhost:7938/icon/' + this.state.session._actions[key].IconName + "/img",
            rank: -100000
        }))


        if (this.state.shown) {
            return [
                <ItemList key="itemlist" style={itemListStyle} items={items} select={this.select} activate={this.execute} dismiss={this.dismiss} />,
                <Indicator key="indicator" />
            ];
        }
        else if (this.state.notifications.length > 0) {
            return <div> {
                this.state.notifications.map(n => {
                    let item = { url: n._self, name: n.Subject, comment: n.Body, image: n.Image && "http://localhost:7938" + n.Image }
                    return <Item key={item.url} item={item} />
                })
            } </div>
        } else {
            return null
        }
    }
}

export { Do }
