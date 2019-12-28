//Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.


import React from 'react'
import { WIN, publish, subscribe, filterAndSort } from "../common/utils";
import { ItemList } from "../common/itemlist"
import { Item } from "../common/item"
import { Indicator } from "./indicator";
import { T } from "../common/translate";
import { monitorUrl, getUrl, postUrl, iconUrl } from '../common/monitor';

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
        super(props);
        [this.model, this.controller] = makeModelAndController();
        ["notifications", "windows", "applications", "session"].forEach(key => this.controller.itemMap.set(key, []))
        monitorNotifications(this.controller);

        this.state = {open: false};
        this.listenForUpDown();
        this.handleBlurEvents();
    };

    focusInput = () => document.getElementById("input") && document.getElementById("input").focus();

    listenForUpDown = () => {
        let that = this;
        http.createServer(function (req, res) {
            that.open();
            if (req.url === "/up") {
                that.controller.move(false);
            } else {
                that.controller.move(true);
            }
            res.end('')
        }).listen("/run/user/1000/org.refude.panel.do");
    };

    handleBlurEvents = () => WIN.on('blur', this.close);


    keyDown = (event) => {
        let { key, ctrlKey, shiftKey, altKey, metaKey } = event;

        if (key === "Tab" && !ctrlKey && shiftKey && !altKey && !metaKey) this.controller.move(false);
        else if (key === "Tab" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.controller.move(true);
        else if (key === "ArrowUp" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.controller.move(false);
        else if (key === "ArrowDown" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.controller.move(true);
        else if (key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.activate();
        else if (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey) this.activate(); 
        else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.close();
        else {
            return;
        }
        event.preventDefault();
    };

    activate = (item) => {
        if (!item) {
            item = this.model.items[this.model.selectedIndex()];
        }
        if (item) {
            postUrl(item.url, response => { this.close() });
        }
    }

    open = () => {
        console.log("open, this.state.open:", this.state.open);
        if (! this.state.open) {
            this.setState({open: true});
            fetchResources(this.controller);
        }
        WIN.focus()
    };

    close = () => {
        console.log("close")
        let input = document.getElementById("input");
        if (input) {
            input.value = "";
            this.controller.setTerm("");
        }
        this.setState({open: undefined});
        this.controller.clear(["windows", "applications", "session"])
    };

    render = () => {

        let outerStyle = {
            maxWidth: "300px",
            maxHeight: "300px",
            display: "flex",
            flexFlow: "column",
            paddingTop: this.state.open ? "0.3em" : "0px",
            paddingLeft: "0.3em",
        };

        let searchBoxStyle = {
            boxSizing: "border-box",
            paddingRight: "5px",
            width: "calc(100% - 16px)",
            marginTop: "4px",
            marginBottom: "6px",
        };

        let inputStyle = {
            width: "100%",
            height: "36px",
            borderRadius: "5px",
            outlineStyle: "none",
        };

        return [ 
            <div onKeyDown={this.keyDown} style={outerStyle}>
                {this.state.open &&
                  <div style={searchBoxStyle}>
                    <input id="input" 
                           style={inputStyle} 
                           type="search" 
                           onChange={e => this.controller.setTerm(e.target.value.toLowerCase())} 
                           autoComplete="off"
                           autoFocus/>
                  </div>}
                <ItemList key="itemlist" model={this.model} onClick={this.controller.select} onDoubleClick={this.activate} />
            </div>,
            <Indicator key="indicator" model={this.model}/>

        ];
    }
}

    let fetchResources = (controller) => {
        getUrl("/windows", resp => {
            controller.itemMap.set("windows", resp.data
                .filter(w => w.States.indexOf("_NET_WM_STATE_ABOVE") < 0)
                .map(w => ({
                    group: T("Open windows"),
                    url: w._self,
                    name: w.Name,
                    comment: "",
                    image: iconUrl(w.IconName),
                    iconStyle: windowIconStyle(w),
                    indicator: {
                        X: w.X, Y: w.Y, W: w.W, H: w.H,
                        ImageUrl: `http://localhost:7938${w._self}/screenshot?downscale=3`
                    }, 
                    matchEmpty: true
                }))
            )
            controller.update();
        });

        getUrl("/applications", resp => {
            controller.itemMap.set("applications", resp.data
                .filter(a => !a.NoDisplay)
                .map(a => ({
                    group: T("Applications"),
                    url: a._self,
                    name: a.Name,
                    comment: a.Comment || '',
                    image: iconUrl(a.IconName),
                    matchFluffy: true
                })));
            controller.update();
        });

        getUrl("/session", resp => {
            controller.itemMap.set("session", Object.keys(resp.data._post)
                .map(key => ({
                    group: T("Leave"),
                    url: resp.data._self + "?action=" + key,
                    name: key,
                    comment: resp.data._post[key].Description,
                    image: iconUrl(resp.data._post[key].IconName),
                    matchFluffy: true
                })));
            controller.update();
        });
    }

let monitorNotifications = (controller) => {
    monitorUrl("/notifications", resp => {
        controller.itemMap.set("notifications", resp.data.map(n => ({
            group: T("Notifications"),
            url: n._self,
            name: n.Subject,
            comment: n.Body,
            image: n.Image ? "http://localhost:7938" + n.Image : iconUrl(n.IconName),
            matchEmpty: true
        })))
        controller.update();
    })
}


let makeModelAndController = () => {
    let term = "";

    let model = {
        updateListeners: [],
        selectedItem: null, 
        items: [],
        selectedIndex: () => model.selectedItem ? model.items.findIndex(i => i.url === model.selectedItem.url) : -1
    }
 
    let notifyListeners = () => {
        model.updateListeners.forEach(l => l());
    }    

    let controller = {
        itemMap: new Map(),
        setTerm: (t) => {
            term = t.toLowerCase();
            controller.update()
        },
        move: (down) => {
            let index = model.selectedIndex()
            if (index > -1) {
                let numItems = model.items.length;
                index = (index + numItems + (down ? 1 : -1)) % numItems;
                model.selectedItem = model.items[index];
                notifyListeners();
            }
        },
        select: (item) => {
            if (model.items.findIndex(i => i.url === item.url) > - 1) {
                model.selectedItem = item;
                notifyListeners();
            } 
        },
        update: () => { 
            model.items = [];
            controller.itemMap.forEach((itemList, k) => {
                model.items.push(...filterAndSort(itemList, term));
            });
            let index = model.selectedIndex();
            if (index > -1) {
                model.selectedItem = model.items[index]
            } else {
                model.selectedItem = model.items[0]
            }
        
            notifyListeners();
        },
        clear: (keys) => {
           keys.forEach(k => controller.itemMap.set(k, []));
           controller.update();
        }

    }

    return [model, controller];
}

export { Do }
