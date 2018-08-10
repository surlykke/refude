// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.


const http = require('http')
import React from 'react'
import {render} from 'react-dom'
import {doSearch, doPost} from '../common/http'
import {WIN, devtools, watchWindowPositionAndSize, showWindowIfHidden, hideWindow} from "../common/nw";
import {TitleBar} from "../common/titlebar";
import {ItemList} from "../common/itemlist"

const searches = [
    {
        group: "Open windows",
        service: "wm-service",
        query: "r.Name neq 'Refude Do' and r.Name neq 'refudeDo'",
        forWindows: true,
    },
    {
        group: "Applications",
        minTermSize: 1,
        service: "desktop-service"
    },
    {
        group: "Leave",
        minTermSize: 1,
        service: "power-service"
    }
];

class Do extends React.Component {
    constructor(props) {
//        devtools();
        super(props);
        this.state = {items: new Map()};
        this.itemList = React.createRef();
        this.windows = [];

        this.listenForUpDown();
        this.handleBlurEvents();
        this.fetchWindowsAndItems();

        WIN.on('loaded', watchWindowPositionAndSize())
    };

    listenForUpDown = () => {
        let outerThis = this;
        http.createServer(function (req, res) {
            if (req.url === "/u") {
                outerThis.showWin();
                outerThis.itemList.current.move(true);
            } else if (req.url === "/d") {
                this.showWin();
                outerThis.itemList.current.move(false);
            }
            res.end('')
        }).listen("/run/user/1000/org.refude.do");
    }

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
        showWindowIfHidden();
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
        if (this.needsInitialize) {
            this.fetchWindowsAndItems();
            this.needsInitialize = undefined;
        }
    };

    select = item => {
        console.log(item._self, "selected");
    };

    execute = (item) => {
        doPost(item).then(response => {
            this.onDismiss();
        })
    };

    onDismiss = () => {
        this.itemList.current.clear();
        this.needsInitialize = true;
        hideWindow()
    };

    render = () => {
        let style = {
            display: "flex",
            flexFlow: "column",
            height: "100%"
        };

        return  <div style={style}>
                    <TitleBar key="titlebar"/>
                    <ItemList key="itemlist"
                              items={this.state.items}
                              select={this.select}
                              execute={this.execute}
                              onDismiss={this.onDismiss}
                              ref={this.itemList}/>
                </div>
    };
}

render(<Do/>, document.getElementById('root'));
