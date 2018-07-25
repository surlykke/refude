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

class Container extends React.Component {
    constructor(props) {
        super(props);
        this.state = {items: []};
        this.itemList = React.createRef();
        this.windows = [];

        nwSetup((argv) => {
            this.readArgs(argv);
            this.windows = [];
        });

        WIN.on('focus', () => {this.hasfocus = true;});
        WIN.on('blur', () => {
            this.hasfocus = false;
            // TAB momentarily unfocuses window - so we wait a bit
            setTimeout(() => { if (!this.hasfocus) this.itemList.current.dismiss(); }, 100);
        });
        // devtools();
    };

    componentDidMount = () => {
        watchPos();
        this.termChange("");
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
            this.setState({items: collected});
        }
    };


    termChange = (newTerm) => {
        this.fetch(newTerm, searches, []);
    };

    select = item => {
        console.log(item._self, "selected");
    };

    execute = (item) => {
        doPost(item).then(response => {
            this.dismiss();
        })
    };

    onDismiss = () => {
        console.log("dismiss");
        nwHide()
    };


    getWindows = () => {
        doSearch("wm-service", "application/vnd.org.refude.wmwindow+json").then(resp => {
            this.windows = resp.json;
            this.termChange("");
        }, resp => {
            this.termChange("");
        });
    };

    readArgs = (args) => {
        adjustPos();
        this.termChange("");
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
                      onTermChange={this.termChange}
                      select={this.select}
                      execute={this.execute}
                      onDismiss={this.onDismiss}
                      ref={this.itemList}/>
        )
    }
}

render(<Container/>, document.getElementById('root'));
