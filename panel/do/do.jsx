//Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.


const http = require('http')
import React from 'react'
import {render} from 'react-dom'
import {doSearch, doPostPath, doGet} from '../../common/http'
import {Utils, WIN, devtools, applicationRank} from "../../common/utils";
import {ItemList} from "../../common/itemlist"
import {T} from "../../common/translate";



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
        this.resources = {windows: [], applications: []};
        this.term = "";
        this.state = {items: []};
        this.itemList = React.createRef();
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


    getResources(service, mimetype, resourceKey) {
        this.resources[resourceKey] = [];
        doSearch(service, mimetype, "").then(resp => {
            this.resources[resourceKey] = resp.json;
            this.filterAndSort()
        });
    }

    getResource(service, path, resourceKey) {
        doGet({service: service, path: path}).then(resp => {
            this.resources[resourceKey] = resp.json;
        });
    }

    filterAndSort = () => {
        let term = this.term.toLowerCase();

        let items = [];
        this.resources.windows
            .filter(w => w.States.indexOf("_NET_WM_STATE_ABOVE") < 0)
            .filter(w => w.Name.toLowerCase().indexOf(term) > -1)
            .sort((w1, w2) => w1.StackOrder - w2.StackOrder)
            .forEach(w => {
                items.push({
                    group: T("Open windows"),
                    url: w._self,
                    description: w.Name,
                    iconName: w._actions['default'].IconName,
                    iconStyle: windowIconStyle(w)
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
            console.log("session:", this.resources.session);
            for (let [id, a] of Object.entries(this.resources.session._actions)) {
                console.log("Consider", id, a.Description);
                if (a.Description.toLowerCase().indexOf(term) > -1) {
                    let item = {
                        group: T("Leave"),
                        url: this.resources.session._self + "?action=" + id,
                        description: a.Description,
                        iconName: a.IconName
                    };
                    console.log("Adding", item)
                    items.push(item);
                }
            }
        }

        this.setState({items: items});
    };

    showWin = () => {
        if (!this.state["shown"]) {
            this.getResources("wm-service", "application/vnd.org.refude.wmwindow+json", "windows")
            this.getResources("desktop-service", "application/vnd.org.refude.desktopapplication+json", "applications")
            this.getResource("power-service", "/session", "session");
            this.setState({"shown": true});
        }
        WIN.focus();
    };

    termChange = term => {
        console.log("termChange:", term);
        this.term = term;
        this.filterAndSort();
    };

    execute = (item) => {
        console.log("execute: ", item);
        doPostPath(item.url).then(response => {
            this.onDismiss();
        })
    };

    onDismiss = () => {
        if (this.state.shown) {
            this.resources = {};
            this.term = "";
            this.setState({items: []});
            this.setState({"shown": undefined});
        }
    };

    render = () => {
        let itemListStyle = {maxWidth: "300px", maxHeight: "300px"};
        if (this.state.shown)
            return <ItemList key="itemlist"
                             style={itemListStyle}
                             items={this.state.items}
                             onTermChange={this.termChange}
                             execute={this.execute}
                             onDismiss={this.onDismiss}
                             onUpdated={this.onUpdated}
                             ref={this.itemList}/>
        else
            return null
    };
}

//render(<Do/>, document.getElementById('root'));
export {Do}