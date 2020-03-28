//Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.


import React from 'react'
import { WIN, publish } from "../common/utils";
import { Indicator } from "./indicator";
import { T } from "../common/translate";
import { getUrl, postUrl, monitorSSE } from '../common/monitor';

const http = require('http');

class Do extends React.Component {
    constructor(props) {
        super(props);
        this.state = { resources: [], open: false, term: "" }
        this.listenForUpDown();
        this.handleBlurEvents();
    };

    componentDidUpdate = () => {
        // Scroll selected item into view
        if (this.state.selected) {
            let selectedDiv = document.getElementById(this.state.selected);
            if (selectedDiv) {
                let listDiv = document.getElementById("itemListDiv");
                if (listDiv) {
                    let { top: listTop, bottom: listBottom } = listDiv.getBoundingClientRect();
                    let { top: selectedTop, bottom: selectedBottom } = selectedDiv.getBoundingClientRect();
                    if (selectedTop < listTop) listDiv.scrollTop -= (listTop - selectedTop + 25)
                    else if (selectedBottom > listBottom) listDiv.scrollTop += (selectedBottom - listBottom + 10)
                }
            }
        }
        publish("componentUpdated");
    };

    listenForUpDown = () => {
        let that = this;
        http.createServer(function (req, res) {
            that.open(() => { that.move(req.url !== "/up") });
            res.end('')
        }).listen("/run/user/1000/org.refude.panel.do");
    };

    handleBlurEvents = () => WIN.on('blur', this.close);

    updateResourceList = () => {
        let url = `/search/desktop?term=${this.state.term}`
        getUrl(url, resp => {
            this.setState({ resources: resp.data }, this.ensureSelection)
        })
    }

    ensureSelection = () => {
        if (!this.selectedResource()) {
            this.setState({ selected: this.state.resources[0] && this.state.resources[0].Self })
        }
    }

    selectedResource = () => {
        console.log("selectedREsource looking for", this.state.selected)
        let res = this.state.selected && this.state.resources.find(r => r.Self === this.state.selected)
        console.log("Returning", res)
        return res
    }

    setTerm = (term) => {
        this.setState({ term: term.toLowerCase() }, this.updateResourceList)
    }

    keyDown = (event) => {
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
        let index = this.state.resources.findIndex(r => r.Self === this.state.selected)
        if (index > -1) {
            index = (index + this.state.resources.length + (down ? 1 : -1)) % this.state.resources.length;
            this.setState({ selected: this.state.resources[index].Self })
            //notifyListeners();
        }
    }

    select = (url) => {
        this.setState({ selected: url })
    }

    activate = (url) => {
        url = url || this.state.selected
        if (url) {
            postUrl(url, response => { this.close() });
        }
    }

    open = (callback) => {
        this.setState({ open: true, term: "" }, this.updateResourceList)
        callback && callback()
        publish("doOpen")
        WIN.focus()
    };

    close = () => {
        this.setState({ open: false, resources: [], term: "" });
        publish("doClose")
    };

    render = () => {

        let doStyle = {
            maxWidth: "300px",
            maxHeight: "300px",
            display: "flex",
            flexFlow: "column",
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
        this.state.resources.forEach(r => {
            if (prevType !== r.Type) {
                items.push(<div style={headingStyle}>{r.Type}</div>)
            }
            items.push(
                <div id={r.Self} style={itemStyle(r.Self)} onClick={() => this.select(r.Self)} onDoubleClick={() => this.activate(r.Self)}>
                    <img width="24px" height="24px" style={iconStyle(r)} src={iconUrl(r)} alt="" />
                    <div style={nameStyle}>{r.Title}</div>
                    <div style={commentStyle}>{r.Comment}</div>
                </div>
            )
            prevType = r.Type
        })

        if (this.state.open) {
            return <>
            <div style={doStyle}>
                <div style={searchBoxStyle} onKeyDown={this.keyDown}>
                    <input id="input"
                        style={inputStyle}
                            type="search"
                            onChange={e => this.setTerm(e.target.value)}
                            value={this.state.term}
                            autoComplete="off"
                            autoFocus />
                </div>
                <div id="itemListDiv" style={innerStyle}>
                    {items}
                </div>
            </div>
            <Indicator key="indicator" res={this.selectedResource()} />
            </>
        } else {
            return null;
        }
    }
}

export { Do }
