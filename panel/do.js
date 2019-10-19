//Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.


import React from 'react'
import { WIN, publish, subscribe} from "../common/utils";
import { ItemList } from "../common/itemlist"
import { Item } from "../common/item"
import { Indicator } from "./indicator";
import { T } from "../common/translate";
import { monitorUrl, getUrl, postUrl, iconUrl } from '../common/monitor';

const http = require('http');

let rank = (name, comment, lowercaseTerm) => {
    let tmp = name.toLowerCase().indexOf(lowercaseTerm)
    if (tmp > -1) {
        console.log("rank('" + name + "', '" + comment + "', '" + lowercaseTerm + "') returning", tmp)
        return tmp
    }
    if (comment) {
        tmp = comment.toLowerCase().indexOf(lowercaseTerm)
        if (tmp > -1) {
            console.log("rank('" + name + "', '" + comment + "', '" + lowercaseTerm + "') returning", tmp + 100)
            return 100 + tmp
        }
    }
    return -1
};

let match = (name, comment, lowercaseTerm) => rank(name, comment, lowercaseTerm) > -1

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
    notifications.forEach(n => t = !(t < timeToExpiry(n)) ? timeToExpiry(n) : t)
    return t
}

class Do extends React.Component {
    constructor(props) {
        //devtools();
        super(props);
        this.state = { notifications: [], windows: [], applications: [], session: { _post: {} }, term: "" };

        this.listenForUpDown();
        this.handleBlurEvents();

        subscribe("termChanged", newTerm => this.setState({ term: newTerm.toLocaleLowerCase() }))
        subscribe("itemSelected", url => publish("windowSelected", this.state.windows.find(w => w._self === url)))
        subscribe("itemActivated", item => postUrl(item.url, response => { this.reset() }))
        subscribe("dismiss", this.dismiss)
    };

    componentDidMount = () => {
        monitorUrl("/notifications", resp => {
            if (this.state.notifications.length === 0) {
                this.notificationFilterScheduler(resp.data)
            } else { // scheduler will already be running
                this.setState({ notifications: resp.data.filter(n => timeToExpiry(n) > 0) })
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
                if (!this.hasfocus) this.dismiss(true);
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


    dismiss = (blur) => {
        if (this.state.shown && !blur) {
            // Return focus to whatever was at the top of the (normal) stack
            this.state.windows[0] && postUrl(this.state.windows[0]._self)
        }
        this.reset()
    };

    reset = () => {
        this.setState({ "shown": undefined, term: "" })
        publish("reset")
    }

    render = () => {
        let itemListStyle = { maxWidth: "300px", maxHeight: "300px" };
        let term = this.state.term
        let items = []
        this.state.notifications.
            filter(n => match(n.Subject, n.Body, term)).
            forEach(n => {
                items.push({
                    group: T("Notifications"),
                    url: n._self,
                    name: n.Subject,
                    comment: n.Body,
                    image: n.Image ? "http://localhost:7938" + n.Image : iconUrl(n.IconName)
                })
            })

        this.state.windows.
            filter(w => match(w.Name, undefined, term)).
            forEach(w => items.push({
                group: T("Open windows"),
                url: w._self,
                name: w.Name,
                comment: "",
                image: iconUrl(w.IconName),
                iconStyle: windowIconStyle(w),
            }))

        if (term !== '') {
            this.state.applications.
                filter(a => match(a.Name, a.Comment, term)).
                sort((a1, a2) => rank(a1.Name, a1.Comment, term) - rank(a2.Name, a2.Comment, term)).
                forEach(a => items.push({
                    group: T("Applications"),
                    url: a._self,
                    name: a.Name,
                    comment: a.Comment || '',
                    image: iconUrl(a.IconName)
                }))

            let desc = key => this.state.session._post[key].Description
            Object.keys(this.state.session._post).
                filter(key => match(key, desc(key), term)).
                sort((k1, k2) => rank(k1, desc(k1), term) - rank(k2, desc(k2), term)).
                forEach(key => items.push({
                    group: T("Leave"),
                    url: this.state.session._self + "?action=" + key,
                    name: key,
                    comment: this.state.session._post[key].Description,
                    image: iconUrl(this.state.session._post[key].IconName),
                }))
        }

        if (this.state.shown) {
            return [
                <ItemList key="itemlist" style={itemListStyle} items={items} />,
                <Indicator key="indicator" />
            ];
        } else if (this.state.notifications.length > 0) {
            return <div> {
                this.state.notifications.map(n => {
                    let item = {
                        url: n._self,
                        name: n.Subject,
                        comment: n.Body,
                        image: n.Image ? "http://localhost:7938" + n.Image : iconUrl(n.IconName)
                    }
                    return <Item key={item.url} item={item} />
                })
            } </div>
        } else {
            return null
        }
    }
}

export { Do }
