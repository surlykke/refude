//Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.


const http = require('http')
import React from 'react'
import {render} from 'react-dom'
import {doSearch, doPostPath, doGet} from '../../common/http'
import {NW, WIN, devtools} from "../../common/nw";
import {ItemList} from "../../common/itemlist"
import {T} from "../../common/translate";


let rankApplication = (app, lowercaseTerm) => {
    if (lowercaseTerm !== "") {
        let tmp = app.Name.toLowerCase().indexOf(lowercaseTerm);
        if (tmp > -1) {
            return -tmp
        }
        if (app.Comment) {
            tmp = app.Comment.toLowerCase().indexOf(lowercaseTerm);
            if (tmp > -1) {
                return -tmp - 100
            }
        }
    }
    return 1;
};

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
        this.term = ""
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

    filterAndSort = () => {
        let term = this.term.toLowerCase();

        let items = [];
        this.resources.windows
            .filter(w => w.States.indexOf("_NET_WM_STATE_ABOVE") < 0)
            .filter(w => w.Name.toLowerCase().indexOf(term) > -1)
            .sort((w1, w2) => w1.StackOrder - w2.StackOrder)
            .forEach(w => {
                items.push({
                    Group: T("Open windows"),
                    Self: w._self,
                    Name: w.Name,
                    Actions: w._actions,
                    IconUrl: w.IconUrl,
                    IconStyle: windowIconStyle(w)
                });
            });

        this.resources.applications.forEach(a => a.__rank = rankApplication(a, term));
        this.resources.applications
            .filter(a => a.__rank < 1)
            .sort((a1, a2) => (a2.__rank - a1.__rank))
            .forEach(a => {
                items.push({
                    Group: T("Applications"),
                    Self: a._self,
                    Name: a.Name + ' - ' + a.Comment,
                    Actions: a._actions,
                    IconUrl: a.IconUrl
                })
            });
        if (term.length > 0 && this.session && this.session._actions["default"].Description.toLowerCase().indexOf(term) > -1) {
           items.push({
                Group: T("Leave"),
                Self: this.session._self,
                Name: "Leave",
                Actions: this.session._actions,
                IconUrl: `http://localhost:7938/icon-service/icon?name=${this.session._actions['default'].IconName}`
            });
        }
        this.setState({items: items});
    };

    showWin = () => {
        if (!this.state["shown"]) {
            this.getResources("wm-service", "application/vnd.org.refude.wmwindow+json", "windows")
            this.getResources("desktop-service", "application/vnd.org.refude.desktopapplication+json", "applications")
            doGet({service: "power-service", path: "/session"}).then(resp => {
                console.log("setting session:", resp.json);
                this.session = resp.json;
                this.filterAndSort();
            })
            this.setState({"shown": true});
        }
        WIN.focus();
    };

    termChange = term => {
        console.log("termChange:", term);
        this.term = term;
        this.filterAndSort();
    }

    execute = (item) => {
        console.log("execute: ", item);
        doPostPath(item.Self).then(response => {
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