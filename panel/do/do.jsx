//Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.


const http = require('http')
import React from 'react'
import {render} from 'react-dom'
import {doSearch, doPost, doPatch} from '../../common/http'
import {NW, WIN, devtools} from "../../common/nw";
import {ItemList} from "../../common/itemlist"
import {T} from "../../common/translate";
import {showSelectedWindow} from "./indicate";

const searches = [
    {
        group: T("Notifications"),
        service: "notifications-service",
    },
    {
        group: T("Open windows"),
        service: "wm-service",
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
        doSearch("wm-service", "application/vnd.org.refude.wmwindow+json").then(resp => {
            this.windows = {}
            resp.json.forEach(win => this.windows[win._self] = win);
            this.fetchItems();
        }, resp => {
            this.fetchItems();
        });
    };

    fetchItems = () => {
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
        if (!this.state["shown"]) {
            this.fetchWindowsAndItems();
            this.setState({"shown": true});
        }
        WIN.focus();
    };

    select = item => {
        let window;
        if (item && item._relates && item._relates["application/vnd.org.refude.wmwindow+json"]) {
            window = this.windows["/wm-service" + item._relates["application/vnd.org.refude.wmwindow+json"][0]];
        }

        if (window) {
            showSelectedWindow(window);
        } else {
            showSelectedWindow(null);
        }
    };

    execute = (item) => {
        doPost(item).then(response => {
            this.onDismiss();
        })
    };

    onDismiss = () => {
        if (this.state.shown) {
            this.select(false);
            this.itemList.current.clear();
            this.needsInitialize = true;
            this.setState({"shown": undefined});
        }
    };

    render = () => {
        let style = {
            flexFlow: "column",
            maxHeight: "600px",
            width: "300px"
        };


        if (this.state.shown)
            return <div style={style}>
                <ItemList key="itemlist"
                          items={this.state.items}
                          select={this.select}
                          execute={this.execute}
                          onDismiss={this.onDismiss}
                          onUpdated={this.onUpdated}
                          ref={this.itemList}/>
            </div>
        else
            return null
    };
}

//render(<Do/>, document.getElementById('root'));
export {Do}