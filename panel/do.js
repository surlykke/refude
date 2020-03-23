//Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.


import React from 'react'
import { WIN, publish } from "../common/utils";
import { Indicator } from "./indicator";
import { T } from "../common/translate";
import { getUrl, postUrl, monitorSSE} from '../common/monitor';

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
        this.state = {resources: [], open: false, term: ""}
        this.listenForUpDown();
        monitorSSE("events", () => this.state.open || this.updateResourceList(), this.updateResourceList)
        this.handleBlurEvents();
    };

    componentDidUpdate = () => {
        // Scroll selected item into view
        if (this.state.selected) {
            let selectedDiv = document.getElementById(this.state.selected);
            if (selectedDiv) {
                let listDiv = document.getElementById("itemListDiv");
                let { top: listTop, bottom: listBottom } = listDiv.getBoundingClientRect();
                let { top: selectedTop, bottom: selectedBottom } = selectedDiv.getBoundingClientRect();
                if (selectedTop < listTop) listDiv.scrollTop -= (listTop - selectedTop + 25)
                else if (selectedBottom > listBottom) listDiv.scrollTop += (selectedBottom - listBottom + 10)
            }
        }
        publish("componentUpdated");
    };

    focusInput = () => document.getElementById("input") && document.getElementById("input").focus();

    listenForUpDown = () => {
        let that = this;
        http.createServer(function (req, res) {
            that.open(() => {that.move(req.url !== "/up")});
            res.end('')
        }).listen("/run/user/1000/org.refude.panel.do");
    };

    handleBlurEvents = () => WIN.on('blur', this.close);

    updateResourceList = () => {
        console.log("updateResourceList")
        if (this.state.open) {
            let url = `/search/desktop?term=${this.state.term}` 
            console.log("Getting", url)
            getUrl(url, resp => {
                this.setState({resources: resp.data}, this.ensureSelection)
            })
        } else {
            getUrl("/search/events", resp => {
                this.setState({resources: resp.data, selected: undefined})
            }) 
        }
    }

    ensureSelection = () => {
        if (!(this.state.selected && this.state.resources.find(r => r.Self === this.state.selected))) {
            let firstRes = this.state.resources[0]
            this.setState({selected: firstRes && firstRes.Self})
        }
    }

    setTerm = (term) => {
        this.setState({term: term.toLowerCase()}, this.updateResourceList)
    }

    keyDown = (event) => {
        console.log("keyDown: ", event)
        let { key, ctrlKey, shiftKey, altKey, metaKey } = event;

        if (key === "Tab" && !ctrlKey && shiftKey && !altKey && !metaKey) this.move(false);
        else if (key === "Tab" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true);
        else if (key === "ArrowUp" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(false);
        else if (key === "ArrowDown" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true);
        else if (key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.activate();
        else if (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey) this.activate();
        else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.close();
        else {
            return;
        }
        event.preventDefault();
    };

    move = (down) => {
        if (!this.state.open) return;
        console.log("move", down, "selected just now:", this.state.selected)
        let index = this.state.resources.findIndex(r => r.Self === this.state.selected)
        console.log("index:", index)
        if (index > -1) {
            index = (index + this.state.resources.length + (down ? 1 : -1)) % this.state.resources.length;
            console.log("Set selected:", this.state.resources[index].Self)
            this.setState({selected: this.state.resources[index].Self}) 
            //notifyListeners();
        }
    }

    select = (url) => {
        this.setState({selected: url})
    }

    activate = (url) => {
        url = url || this.state.selected 
        if (url) {
            postUrl(url, response => { this.close() });
        }
    }

    open = (callback) => {
        console.log("open, this.state.open:", this.state.open);
        this.setState({open: true, term: ""}, this.updateResourceList)
        callback && callback()
        WIN.focus()
    };

    close = () => {
        console.log("close")
        this.setState({open: false, term: ""}, this.updateResourceList);
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

        let innerStyle = {
            overflowY: "scroll"
        };

        let headingStyle = {
            fontSize: "0.9em",
            color: "gray",
            fontStyle: "italic",
            marginTop: "5px",
            marginBottom: "3px",
        };

        let itemStyle = (self) => {
            let style = {
                marginRight: "5px",
                padding: "4px",
                verticalAlign: "top",
                overflow: "hidden",
                height: "30px",
            };

            if (this.state.selected === self) {
                    Object.assign(style, {
                        border: "solid black 2px",
                        borderRadius: "5px",
                        boxShadow: "1px 1px 1px #888888",
                })
            }
            return style
        }

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

        let nameStyle = {
            overflow: "hidden",
            whiteSpace: "nowrap",
            marginRight: "6px",
        };
    
        let commentStyle = {
            fontSize: "0.8em",
        };


        let iconStyle = (res) => {
            console.log("iconStyle, Type:", res.Type, "Data.States:", res.Data.States)
            let style = {
                float: "left",
                marginRight: "6px"
            };
           
            if (res.Type === "window") {
                Object.assign(style, {
                    WebkitFilter: "drop-shadow(5px 5px 3px grey)",
                    overflow: "visible"
                });
                
                if (res.Data.States && res.Data.States.includes("_NET_WM_STATE_HIDDEN")) {
                    Object.assign(style, {
                        marginLeft: "10px",
                        marginTop: "10px",
                        width: "14px",
                        height: "14px",
                        opacity: "0.7"
                    })
                }
            }

            return style
        }

        let iconUrl = (res) => {
            if (res.IconName) {
                return `http://localhost:7938/icon?name=${res.IconName}&theme=oxygen`
            } else {
                return ""
            }
        }

        let prevType 

        let items = []
        this.state.resources.forEach( r => {
            if (prevType !== r.Type) {
                items.push(<div style={headingStyle}>{r.Type}</div>)
            }
            items.push(
                <div id={r.Self} style={itemStyle(r.Self)} onClick={() => this.select(r.Self)} onDoubleClick={() => this.activate(r.Self)}>
                    <img width="24px" height="24px" style={iconStyle(r)} src={iconUrl(r)} alt=""/>
                    <div style={nameStyle}>{r.Title}</div>
                    <div style={commentStyle}>{r.Comment}</div>
                </div>
            )
            prevType = r.Type
        })

        return [
            <div onKeyDown={this.keyDown} style={outerStyle}>
                {this.state.open &&
                    <div style={searchBoxStyle}>
                        <input id="input"
                            style={inputStyle}
                            type="search"
                            onChange={e => this.setTerm(e.target.value)}
                            value={this.state.term}
                            autoComplete="off"
                            autoFocus />
                    </div>}
                <div id="itemListDiv" style={innerStyle}>
                    {items}
                </div>
            </div>,
            <div/>
            //<Indicator key="indicator" model={this.model} />
        ];
    }
}

export { Do }
