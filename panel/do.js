//Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.


import React from 'react'
import { WIN, applicationRank, publish, subscribe } from "../common/utils";
import { ItemList} from "../common/itemlist"
import { Item} from "../common/item"
import { Indicator } from "./indicator";
import { T } from "../common/translate";
import { doGet, doPost, monitorUrl } from '../common/monitor';
const axios = require('axios')
axios.defaults.baseURL = "http://localhost:7938"



const windowSearch = "/windows?q=" + encodeURIComponent('not r.States[%] eq _NET_WM_STATE_ABOVE')
const applicationSearch = "/applications?q=" + encodeURIComponent('not r.NoDisplay eq true')

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

class Do extends React.Component {
    constructor(props) {
        //devtools();
        super(props);
        this.resources = { notifications: [], windows: [], applications: [] };
        this.term = "";
        this.state = { items: [], recentNotificationItems: [] };
        this.display = { x: 0, y: 0, w: 100, h: 100 };

        this.listenForUpDown();
        this.handleBlurEvents();
    };

    componentDidMount = () => {
        subscribe("termChanged", this.termChange);
        subscribe("itemLaunched", this.execute);
        subscribe("dismiss", this.onDismiss);

        let path = () => "/notifications?q=" + encodeURIComponent("r.Created gt " + (new Date().getTime() - 3000)) 

        monitorUrl(path, resp => {
            let recentNotificationItems = resp.data.map(n => {
                return {
                    group: T("Notifications"),
                    url: n._self,
                    description: n.Subject,
                    Comment: n.Body
                }
            });
            this.setState({ recentNotificationItems: recentNotificationItems })
        });
    };


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
        this.resources.notifications
            .filter(n => n.Subject.toLowerCase().indexOf(term) > -1 || n.Body.toLowerCase().indexOf(term) > -1)
            .forEach(n => {
                items.push({
                    group: T("Notifications"),
                    url: n._self,
                    description: n.Subject,
                    Comment: n.Body
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
                    iconName: w._actions['default'].IconName,
                    iconStyle: windowIconStyle(w),
                    bounds: { X: w.X, Y: w.Y, W: w.W, H: w.H }
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
                        iconName: a.IconName,
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
                        iconName: a.IconName
                    };
                    items.push(item);
                }
            }
        }

        this.setState({ items: items });
    };

    showWin = () => {
        if (!this.state["shown"]) {
            this.resources["windows"] = this.resources["applications"] = this.resources["session"] = [];

            doGet("/notifications", resp => {
                this.resources.notifications = resp.data
                this.filterAndSort
            });

            doGet(windowSearch, resp => {
                this.resources.windows = resp.data;
                this.filterAndSort()
            });

            doGet(applicationSearch, resp => {
                this.resources.applications = resp.data;
                this.filterAndSort()
            });
            doGet("/session", resp => {
                this.resources.session = resp.data
            });
            this.setState({ "shown": true });
        }
        WIN.focus();
    };

    termChange = term => {
        this.term = term;
        this.filterAndSort();
    };

    execute = (item) => {
        doPost(item.url, response => {
            this.onDismiss();
        })
    };

    onDismiss = () => {
        if (this.state.shown) {
            this.resources.windows = [];
            this.resources.applications = [];
            this.term = "";
            this.setState({ items: [] });
            this.setState({ "shown": undefined });
        }
    };

    render = () => {
        let itemListStyle = { maxWidth: "300px", maxHeight: "300px" };
        if (this.state.shown) {
            return [
                <ItemList key="itemlist"
                    style={itemListStyle}
                    items={this.state.items}
                    ref={this.itemList} />,
                <Indicator key="indicator" />

            ];
        }
        else if (this.state.recentNotificationItems.length > 0) {
            let content = []
            this.state.recentNotificationItems.forEach(n => {
                content.push(<Item key={n.url} item={n}/>)
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