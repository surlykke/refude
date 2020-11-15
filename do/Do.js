//Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.

import React from 'react'
import ReactDOM from 'react-dom'
import { postUrl, deleteUrl, addParam, getUrl, iconUrl, findLink, path2Url } from '../common/monitor';
import { ipcRenderer } from 'electron'
import "./Do.css"
import "../common/common.css"

export class Do extends React.Component {
    constructor(props) {
        super(props)
        this.history = []
        this.state = {url: "/search/desktop", index: 0, term: ""}
        ipcRenderer.on("doShow", this.fetch)
        ipcRenderer.on("doMove", (evt, up) => up ? this.up() : this.down())
    };


    componentDidUpdate = () => {
        // Scroll selected item into view
        let selectedDiv = this.curLink() && this.curLink().href && document.getElementById(this.curLink().href)
        if (selectedDiv) {
            let listDiv = document.getElementById("itemlistDiv");
            if (listDiv) {
                let { top: listTop, bottom: listBottom } = listDiv.getBoundingClientRect();
                let { top: selectedTop, bottom: selectedBottom } = selectedDiv.getBoundingClientRect();
                if (selectedTop < listTop) listDiv.scrollTop -= (listTop - selectedTop + 25)
                else if (selectedBottom > listBottom) listDiv.scrollTop += (selectedBottom - listBottom + 10)
            }
        }

        ipcRenderer.send("doLinkSelected", this.curLink()) // TODO Optimize
    };

    componentDidMount = () => {
        document.addEventListener("keydown", this.keydownHandler)
    }


    fetch = term => {
        getUrl(addParam(this.state.url, "term", this.state.term), resp => {
            this.setState({
                resource: resp.data, 
                links: resp.data._links.filter(l => l.rel !== "self"), 
                self: findLink(resp.data, "self")
            })
        })
    }

    curLink = () => {
        return this.state.links && this.state.links[this.state.index]
    }

    keydownHandler = (event) => {
        let { key, ctrlKey, shiftKey, altKey, metaKey } = event;

        if ((key === "Tab" && !ctrlKey && shiftKey && !altKey && !metaKey) ||
            (key === "ArrowUp" && !ctrlKey && !shiftKey && !altKey && !metaKey) ||
            (key === "k" && ctrlKey && !shiftKey && !altKey && !metaKey)) {
            this.up()
        } else if ((key === "Tab" && !ctrlKey && !shiftKey && !altKey && !metaKey) ||
            (key === "ArrowDown" && !ctrlKey && !shiftKey && !altKey && !metaKey) ||
            (key === "j" && ctrlKey && !shiftKey && !altKey && !metaKey)) {
            this.down()
        } else if ((key === "ArrowLeft" && !ctrlKey && !shiftKey && !altKey && !metaKey) ||
            (key === "h" && ctrlKey && !shiftKey && !altKey && !metaKey)) {
            this.goBack()
        } else if ((key === "ArrowRight" && !ctrlKey && !shiftKey && !altKey && !metaKey) ||
            (key === "l" && ctrlKey && !shiftKey && !altKey && !metaKey)) {
            this.go()
        } else if ((key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) ||
            (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey)) {
            this.activate()
        } else if (key === "Delete" && !ctrlKey && !shiftKey && !altKey && !metaKey) {
            this.delete()
        } else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) {
            this.dismiss()
        } else {
            return;
        }
        event.preventDefault();
    } 

    up = () => this.select(this.state.index - 1)

    down = () => this.select(this.state.index + 1)

    activate = () => {
        this.curLink() && postUrl(this.curLink().href, this.dismiss)
    }

    delete = () => {
        this.curLink() && deleteUrl(this.curLink().href, this.dismiss)
    }

    go = () => {
        if (this.curLink() && this.curLink().rel === "related") {
            this.history.unshift({url: this.state.url, term: this.state.term, index: this.state.index})
            this.setState({url: this.curLink().href, index: 0, term: ""}, this.fetch)
        }
    }

    goBack = () => {
        let tmp = this.history.shift()
        if (tmp) {
            this.setState({url: tmp.url, term: tmp.term, index: tmp.index}, this.fetch)
        }
    }


    dismiss = () => {
        this.setState({url: "/search/desktop", term: "", index: 0, resource: undefined}, () => ipcRenderer.send("dismiss"))
    }
    
    select = i => {
        let {links} = this.state
        links && links.length > 0 && this.setState({index: (i + links.length)%links.length})
    }

    selectAndActivate = i => {
        let {links} = this.state
        links && links.length > 0 && this.setState({index: (i + links.length)%links.length}, this.activate)
    }
    
    setTerm = term => {
        this.setState({term: term}, this.fetch)
    }

    
    render = () => {
            
        let {resource, links, self} = this.state
        let className = i => i === this.state.index ? "item selected" : "item"
        let iconClassName = link => {
            if (link.profile === "/profile/window") {
                if (link.meta && link.meta.state === "minimized") {
                    return "icon window minimized"
                } else {
                    return "icon window"
                }
            } else {
                return "icon"
            }
        }

        return resource ?
            <>
                <div className="topbar">
                {(!self || links.length > 5) && 
                        <input id="input"
                            className="searchinput"
                            type="search"
                            onChange={e => this.setTerm(e.target.value)}
                            value={this.state.term}
                            autoComplete="off"
                            autoFocus />
                }
                </div>
                {self &&
                <div key="resource" id={self.href} className="item">
                    <img width="24px" height="24px" className={iconClassName(self)} src={path2Url(self.icon)} alt="" />
                    <div className="name">{self.title}</div>
                </div>}            

                <div className="itemlist" id="itemlistDiv">
                {links && links.map((l, i) => {
                    return <div key={l.href} id={l.href} 
                                className={className(i)} 
                                onClick={()=> this.select(i)} 
                                onDoubleClick={() => this.selectAndActivate(i)}>
                            <img className={iconClassName(l)} src={path2Url(l.icon)} alt="" height="24" width="24" />
                            <div className="title"> {l.title}</div>
                        </div>
                }
                )}
                </div>
            </> :
            <div />
    } 
}

ReactDOM.render(<Do />, document.getElementById('do'))