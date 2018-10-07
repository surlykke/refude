//Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.


const http = require('http')
import React from 'react'
import {render} from 'react-dom'
import {doSearch, doPost, doPatch} from '../../common/http'
import {WIN, devtools} from "../../common/nw";
import {ItemList} from "../../common/itemlist"
import {T} from "../../common/translate";

const searches = [
    {
        group: T("Open windows"),
        service: "wm-service",
        query: "r.Name neq 'Refude Do' and r.Name neq 'refudeDo'",
        forWindows: true,
    },
    {
        group: T("Applications"),
        minTermSize: 1,
        service: "desktop-service"
    },
    {
        group: T("Leave"),
        minTermSize: 1,
        service: "power-service"
    }
];

class Do extends React.Component {
    constructor(props) {
        //devtools();
        super(props);
        this.state = {items: new Map()};
        this.itemList = React.createRef();
        this.windows = [];
        this.onUpdated = props.onUpdated;

        this.listenForUpDown();
        this.handleBlurEvents();

    };

    componentDidUpdate = () => {
        this.onUpdated();
    };

    listenForUpDown = () => {
        let that = this;
        http.createServer(function (req, res) {
            console.log("updown:", new Date().getMilliseconds());
            that.showWin();
            if (req.url === "/up") {
                that.itemList.current.move(false);
            } else {
                that.itemList.current.move(true);
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
                if (!this.hasfocus) this.onDismiss();
            }, 100);
        });
    }


    fetchWindowsAndItems = () => {
        console.log("enter fetchWindowsAndItems", new Date().getMilliseconds());
        doSearch("wm-service", "application/vnd.org.refude.wmwindow+json").then(resp => {
            this.windows = {}
            resp.json.forEach(win => this.windows[win._self] = win);
            this.fetchItems();
        }, resp => {
            this.fetchItems();
        });
    };

    fetchItems = () => {
        console.log("enter fetchItems", new Date().getMilliseconds());
        let items = new Map();
        searches.forEach(search => items.set(search.group, [])); // For ordering
        searches.forEach(search => {
            doSearch(search.service, "application/vnd.org.refude.action+json", search.query).then((resp) => {
                resp.json.forEach(item => {
                    item.__minTermSize = search.minTermSize;
                    if (search.forWindows) {
                        item.__iconStyle = this.windowIconStyle(item)
                    }
                });
                items.set(search.group, resp.json);
                this.setState({items: items});
                console.log("fetchItems done", new Date().getMilliseconds());
            }).catch(e => {
                console.log(e);
            });
        });
    };

    windowIconStyle = item => {
        let style = {
            WebkitFilter: "drop-shadow(5px 5px 3px grey)",
            overflow: "visible"
        };
        let win = this.windows["/wm-service" + item._relates["application/vnd.org.refude.wmwindow+json"]];
        if (win && win.States.includes("_NET_WM_STATE_HIDDEN")) {
            Object.assign(style, {
                marginLeft: "10px",
                marginTop: "10px",
                width: "14px",
                height: "14px",
                opacity: "0.7"
            })
        }

        return style
    }


    showWin = () => {
        console.log("showWin:", new Date().getMilliseconds());
        if (!this.state["shown"]) {
            this.fetchWindowsAndItems();
            this.setState({"shown": true});
        }
        WIN.focus();
        console.log("leave showWin:", new Date().getMilliseconds());
    };

    select = item => {
        let id;

        if (item && item._relates && item._relates["application/vnd.org.refude.wmwindow+json"]) {
            let window = this.windows["/wm-service" + item._relates["application/vnd.org.refude.wmwindow+json"][0]];
            if (window) {
                id = window.Id
            }
        }

        if (id) {
            doPatch({_self: "/wm-service/highlight"}, {WindowId: id});
        } else {
            doPatch({_self: "/wm-service/highlight"}, {WindowId: 0});
        }
    };

    execute = (item) => {
        doPost(item).then(response => {
            this.onDismiss();
        })
    };

    onDismiss = () => {
        this.select();
        this.itemList.current.clear();
        this.needsInitialize = true;
        this.setState({"shown": undefined});
    };

    render = () => {
        let style = {
            display: this.state["shown"] ? "flex" : "none",
            flexFlow: "column",
            height: "100%"
        };

        return <div style={style}>
            <ItemList key="itemlist"
                      items={this.state.items}
                      select={this.select}
                      execute={this.execute}
                      onDismiss={this.onDismiss}
                      ref={this.itemList}/>
        </div>
    };
}

//render(<Do/>, document.getElementById('root'));
export {Do}