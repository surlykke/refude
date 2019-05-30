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

const windowSearch = "/windows?q=" + encodeURIComponent('not r.States[%] eq _NET_WM_STATE_ABOVE')
const applicationSearch = "/applications?q=" + encodeURIComponent('not r.NoDisplay eq true')


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


let filterOutOld = (notificationlist) => {
    let twentySecondsAgo = new Date().getTime() - 20000
    return notificationlist.filter(n => Date.parse(n.Created) > twentySecondsAgo)
}

class Do extends React.Component {
    constructor(props) {
        //devtools();
        super(props);
        this.resources = { notifications: [], windows: [], applications: [] };
        this.term = "";
        this.state = { items: [], notifications: [] };
        this.display = { x: 0, y: 0, w: 100, h: 100 };

        this.listenForUpDown();
        this.handleBlurEvents();
    };

    componentDidMount = () => {
        subscribe("termChanged", this.termChange);
        subscribe("itemLaunched", this.execute);
        subscribe("dismiss", this.onDismiss);

        console.log("monitorUrl /notifications")
        monitorUrl("/notifications", resp => {
            console.log("Something happened to notificatoins")
            this.setState({notifications: filterOutOld(resp.data) })
            this.state.shown && this.filterAndSort()
            this.monitorNotifications()
        });
    };

    notificationsAreMonitored = false
    monitorNotifications = () => {
        console.log("monitorNotifications")
        if (this.notificationsAreMonitored || this.state.notifications.length === 0) {
            return
        }
        this.notificationsAreMonitored = true
        setTimeout(() => {
            this.notificationsAreMonitored = false
            this.setState({notifications: filterOutOld(this.state.notifications)})
            this.monitorNotifications()
        }, 
        1000)
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

    filterAndSort = () => {
        let term = this.term.toLowerCase();

        let items = [];
        this.state.notifications
            .filter(n => n.Subject.toLowerCase().indexOf(term) > -1 || n.Body.toLowerCase().indexOf(term) > -1)
            .forEach(n => {
                items.push({
                    group: T("Notifications"),
                    url: n._self,
                    description: n.Subject,
                    comment: n.Body,
                    image: "http://localhost:7938/" + n.Image 
                });
            });
        this.resources.windows
            .filter(w => w.Name.toLowerCase().indexOf(term) > -1)
            .sort((w1, w2) => w1.StackOrder - w2.StackOrder)
            .forEach(w => {
                items.push({
                    group: T("Open windows"),
                    url: w._self,
                    description: w.Name,
                    image: 'http://localhost:7938/icon/' + w.IconName + "/img",
                    iconStyle: windowIconStyle(w),
                    w: w
                });
            });

        if (term.length > 0) {
            this.resources.applications.forEach(a => a.__rank = applicationRank(a, term));
            this.resources.applications
                .filter(a => a.__rank < 1)
                .sort((a1, a2) => (a2.__rank - a1.__rank))
                .forEach(a => {
                    items.push({
                        group: T("Applications"),
                        url: a._self,
                        description: a.Name + (a.Comment ? ' - ' + a.Comment : ''),
                        image: 'http://localhost:7938/icon/' + a.IconName + "/img"
                    })
                });
        }

        if (term.length > 0 && this.resources.session) {
            for (let [id, a] of Object.entries(this.resources.session._actions)) {
                if (a.Description.toLowerCase().indexOf(term) > -1) {
                    let item = {
                        group: T("Leave"),
                        url: this.resources.session._self + "?action=" + id,
                        description: a.Description,
                        image: 'http://localhost:7938/icon/' + a.IconName + "/img"
                    };
                    items.push(item);
                }
            }
        }

        this.setState({ items: items });
    };

    selectorShown = 0
    showWin = () => {
        if (!this.state["shown"]) {
            this.resources["windows"] = this.resources["applications"] = this.resources["session"] = [];
            getUrl("/windows", resp => {
                this.resources.windows = resp.data.filter(w => w.States.indexOf("_NET_WM_STATE_ABOVE") < 0);
                this.filterAndSort()
            });

            getUrl("/applications", resp => {
                this.resources.applications = resp.data.filter(app => ! app.NoDisplay);
                this.filterAndSort()
            });

            getUrl("/session", resp => {
                this.resources.session = resp.data
            });

            this.setState({ "shown": true });
            this.selectorShown = new Date().getTime()
            publish("selectorShown", this.selectorShown)
        }
        WIN.focus();
    };

    termChange = term => {
        this.term = term;
        this.filterAndSort();
    };

    execute = (item) => {
        postUrl(item.url, response => {
            this.cleanUp();
        })
    };

    onDismiss = () => {
        if (this.state.shown) {
            // Return focus to whatever was at the top of the (normal) stack
            this.resources.windows[0] && postUrl(this.resources.windows[0]._self)

            this.cleanUp()
        }
    };

    cleanUp = () => {
            this.resources.windows = [];
            this.resources.applications = [];
            this.term = "";
            this.setState({ items: [] });
            this.setState({ "shown": undefined });
    }


    render = () => {
        let itemListStyle = { maxWidth: "300px", maxHeight: "300px" };
        let prefetch = this.resources.windows.map(w => {
            let imgUrl = `http://localhost:7938/windmp/${w.Id}?downscale=3&${this.selectorShown}`
            return <link rel="prefetch" href={imgUrl}/>
        }
)
        if (this.state.shown) {
            return [
                ...prefetch, 
                <ItemList key="itemlist"
                    style={itemListStyle}
                    items={this.state.items}
                    ref={this.itemList} />,
                <Indicator key="indicator" />

            ];
        }
        else if (this.state.notifications.length > 0) {
            let content = []
            this.state.notifications.forEach(n => {
                let item = {url: n._self, description: n.Subject, Comment: n.Body, image: "http://localhost:7938" + n.Image}
                content.push(<Item key={item.url} item={item}/>)
            })
            return <div>
                {content}
            </div>

        } else {
            return null
        }
    };
}

export { Do }
