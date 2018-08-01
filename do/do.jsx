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
import {linkItems, ItemList} from "../common/itemlist"

const searches = [
    {
        minTermSize: 0,
        service: "wm-service",
        query: "r.Name ~i '@TERM@' and r.Name neq 'Refude Do' and r.Name neq 'refudeDo'",
        group: "Open windows"
    },
    {
        minTermSize: 1,
        service: "desktop-service",
        query: "r.Name ~i '@TERM@'",
        group: "Applications"
    },
    {
        minTermSize: 1,
        service: "power-service",
        query: "r.Name ~i '@TERM@'",
        group: "Leave"
    }
];

class Do extends React.Component {
    constructor(props) {
        //devtools();
        super(props);
        this.state = {items: []};
        this.itemList = React.createRef();
        this.windows = [];
        let outerThis = this;
        http.createServer(function (req, res) {
            console.log("Url:", "'" + req.url + "'");
            if (req.url === "/u") {
                outerThis.showWin();
                outerThis.itemList.current.move(true);
            } else if (req.url === "/d") {
                this.showWin();
                outerThis.itemList.current.move(false);
            }
            res.end('')
        }).listen("/run/user/1000/org.refude.do");

        this.initialize();

        WIN.on('focus', () => this.hasfocus = true);
        WIN.on('blur', () => {
            this.hasfocus = false;
            // TAB momentarily unfocuses window - so we wait a bit to see if it's for real
            setTimeout(() => {
                if (!this.hasfocus) this.onDismiss();
            }, 100);
        });
        WIN.on('loaded', watchWindowPositionAndSize())
    };

    fetch = (searchTerm, searchList, collected) => {
        if (searchList.length > 0) {
            let [head, ...tail] = searchList;
            if (head.minTermSize <= searchTerm.length) {
                let mimetype = "application/vnd.org.refude.action+json";
                let query = head.query.replace("@TERM@", searchTerm);
                doSearch(head.service, mimetype, query).then((resp) => {
                    resp.json.forEach(item => {
                        item.__group = head.group;
                        collected.push(item);
                    });
                    this.fetch(searchTerm, tail, collected);
                }, (resp) => {
                    this.fetch(searchTerm, tail, collected);
                }).catch(() => {
                    this.fetch(searchTerm, tail, collected);
                });
            } else {
                this.fetch(searchTerm, tail, collected);
            }
        } else {
            linkItems(collected);
            collected.forEach(item => {
                if (item._relates && item._relates["application/vnd.org.refude.wmwindow+json"]) {
                    item.__iconStyle = {
                        WebkitFilter: "drop-shadow(5px 5px 3px grey)",
                        overflow: "visible"
                    };
                    let win = this.windows["/wm-service" + item._relates["application/vnd.org.refude.wmwindow+json"]];
                    if (win && win.States.includes("_NET_WM_STATE_HIDDEN")) {
                        Object.assign(item.__iconStyle, {
                            marginLeft: "10px",
                            marginTop: "10px",
                            width: "14px",
                            height: "14px",
                            opacity: "0.7"
                        })
                    }
                }
            });
            this.setState({items: collected});
        }
    };

    showWin = () => {
        if (this.needsInitialize) {
            this.initialize();
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


    initialize = () => {
        doSearch("wm-service", "application/vnd.org.refude.wmwindow+json").then(resp => {
            this.windows = {}
            resp.json.forEach(win => this.windows[win._self] = win);
            this.fetch("", searches, []);
        }, resp => {
            this.fetch("", searches, []);
        });
        showWindowIfHidden();
    };

    render = () => {
        return [
            <TitleBar/>,
            <ItemList key="itemlist"
                         items={this.state.items}
                         onTermChange={(term) => {
                             console.log("termChange:", term);
                             this.fetch(term, searches, []);
                         }}
                         select={this.select}
                         execute={this.execute}
                         onDismiss={this.onDismiss}
                         ref={this.itemList} />
            ]
    }
}


render(<Do/>, document.getElementById('root'));
