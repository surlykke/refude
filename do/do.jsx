// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import {render} from 'react-dom'
import {NW, WIN, devtools, nwHide, nwSetup, doSearch, doPost, watchPos, adjustPos} from '../common/utils'
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
        super(props);
        this.state = {items: []};
        this.itemList = React.createRef();
        this.windows = [];

        nwSetup((argv) => {
            this.readArgs(argv);
        });
        this.initialize();

        WIN.on('focus', () =>  this.hasfocus = true );
        WIN.on('blur', () => {
            this.hasfocus = false;
            // TAB momentarily unfocuses window - so we wait a bit
            setTimeout(() => {
                if (!this.hasfocus) this.itemList.current.dismiss();
            }, 100);
        });
//        devtools();
    };

    componentDidMount = () => {
        watchPos();
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

    /*    if (item.States) {  // Its a window
        Object.assign(iconStyle, {
            WebkitFilter: "drop-shadow(5px 5px 3px grey)",
            overflow: "visible"
        })

        if (item.States.includes("_NET_WM_STATE_HIDDEN")) {

    }
    }*/

    select = item => {
        console.log(item._self, "selected");
    };

    execute = (item) => {
        doPost(item).then(response => {
            nwHide();
        })
    };

    onDismiss = () => {
        nwHide();
    };


    initialize = () => {
        doSearch("wm-service", "application/vnd.org.refude.wmwindow+json").then(resp => {
            this.windows = {}
            resp.json.forEach(win => this.windows[win._self] = win);
            this.fetch("", searches, []);
        }, resp => {
            this.fetch("", searches, []);
        });
    };

    readArgs = (args) => {
        adjustPos();
        this.initialize();
        if (args.includes("up")) {
            this.itemList.current.move(false)
        }
        else {
            this.itemList.current.move(true)
        }
    };

    render = () => {
        return (
            <ItemList items={this.state.items}
                      onTermChange={(term) => this.fetch(term, searches, [])}
                      select={this.select}
                      execute={this.execute}
                      onDismiss={this.onDismiss}
                      ref={this.itemList}
            />
        )
    }
}

render(<Do/>, document.getElementById('root'));
