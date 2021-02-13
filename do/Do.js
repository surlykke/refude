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
        this.state = { url: "/search/desktop", links: [], term: "" }
        ipcRenderer.on("doShow", () => {this.shown = true; this.fetch()})
        ipcRenderer.on("/search/desktop", () => this.shown && this.fetch())
        ipcRenderer.on("doMove", (evt, up) => up ? this.up() : this.down())
    };


    componentDidUpdate = () => {
        // Scroll selected item into view
        let selectedDiv = this.state.curLink && this.state.curLink.href && document.getElementById(this.state.curLink.href)
        if (selectedDiv) {
            let listDiv = document.getElementById("itemlistDiv");
            if (listDiv) {
                let { top: listTop, bottom: listBottom } = listDiv.getBoundingClientRect();
                let { top: selectedTop, bottom: selectedBottom } = selectedDiv.getBoundingClientRect();
                if (selectedTop < listTop) listDiv.scrollTop -= (listTop - selectedTop + 25)
                else if (selectedBottom > listBottom) listDiv.scrollTop += (selectedBottom - listBottom + 10)
            }
        }

    };

    componentDidMount = () => {
        document.addEventListener("keydown", this.keydownHandler)
    }


    fetch = () => {
        getUrl(addParam(this.state.url, "term", this.state.term), resp => {
            let curLink = this.state.curLink
            this.setState({
                resource: resp.data,
                links: resp.data._links.filter(l => l.rel !== "self"),
                self: findLink(resp.data, "self"),
            }, () => this.select(curLink))
        })
    }

    keydownHandler = (event) => {
        let { key, ctrlKey, shiftKey, altKey, metaKey } = event;

        if ((key === "Tab" && shiftKey && !altKey && !metaKey) ||
            (key === "ArrowUp" && !shiftKey && !altKey && !metaKey) ||
            (key === "k" && ctrlKey && !shiftKey && !altKey && !metaKey)) {
            this.up()
        } else if ((key === "Tab" && !shiftKey && !altKey && !metaKey) ||
            (key === "ArrowDown" && !shiftKey && !altKey && !metaKey) ||
            (key === "j" && ctrlKey && !shiftKey && !altKey && !metaKey)) {
            this.down()
        } else if ((key === "ArrowLeft" && !shiftKey && !altKey && !metaKey) ||
            (key === "h" && ctrlKey && !shiftKey && !altKey && !metaKey)) {
            this.goBack()
        } else if ((key === "ArrowRight" && !shiftKey && !altKey && !metaKey) ||
            (key === "l" && ctrlKey && !shiftKey && !altKey && !metaKey)) {
            this.go()
        } else if ((key === "Enter" && !shiftKey && !altKey && !metaKey) ||
            (key === " " && !shiftKey && !altKey && !metaKey)) {
            this.activate(ctrlKey)
        } else if (key === "Delete" && !shiftKey && !altKey && !metaKey) {
            this.del(ctrlKey)
        } else if (key === "Escape" && !shiftKey && !altKey && !metaKey) {
            this.dismiss()
        } else {
            return;
        }
        event.preventDefault();
    }

    index = () => {
        let result 
        if (this.state.curLink) {
            result =  this.state.links.findIndex(l => {
                return this.state.curLink.href === l.href
            })
        }
        return result
    }

    up = () => this.move(-1)
    down = () => this.move(1)

    move = step => {
        if (!this.state.curLink) {
            return
        }
        let i = this.index();
        if (i > -1) {
            this.select(this.state.links[(i + step + this.state.links.length) % this.state.links.length])
        } else {
            this.select()
        }
    }


    activate = (keep) => {
        postUrl("/window/unhighlight") 
        this.state.curLink && postUrl(this.state.curLink.href, keep ? undefined : this.dismiss)
    }

    del = (keep) => {
        this.state.curLink && deleteUrl(this.state.curLink.href, keep ? undefined : this.dismiss)
    }

    go = () => {
        if (this.state.curLink && this.state.curLink.rel === "related") {
            this.history.unshift({ url: this.state.url, term: this.state.term, curLink: this.state.curLink })
            let curLink = this.state.curLink
            this.setState({ url: curLink.href, curLink: undefined, term: "" }, this.fetch)
        }
    }

    goBack = () => {
        let tmp = this.history.shift()
        if (tmp) {
            this.setState({ url: tmp.url, term: tmp.term, curLink: tmp.curLink }, this.fetch)
        }
    }


    dismiss = () => {
        postUrl("/window/unhighlight")
        this.setState({ url: "/search/desktop", term: "", links: [], curLink: undefined, resource: undefined }, 
            () => {this.shown = undefined; ipcRenderer.send("dismiss")})
    }


    select = (link) => {
        if (!link || !this.state.links.find(l => l.href === link.href)) {
            link = this.state.links[0]
        }
      
        if (link && !(this.state.curLink && link.href === this.state.curLink.href)) {
            if (link.profile === "/profile/window") {
                postUrl(link.href + "?action=highlight")
            } else if (!(this.state.self && ("/profile/window" === this.state.self.profile))) {
                postUrl("/window/unhighlight")
            }
        }
        this.setState({curLink: link})
    }

    selectAndActivate = link => {
        this.select(link)
        this.activate()
    }

    setTerm = term => {
        this.setState({ term: term }, this.fetch)
    }


    render = () => {

        let { resource, links, self } = this.state
        let className = link => this.state.curLink && this.state.curLink.href == link.href ? "item selected" : "item"
        let iconClassName = link => {
            if (link.profile === "/profile/window") {
                if (link.hints && link.hints.states && link.hints.states.indexOf("HIDDEN") > -1) {
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
                    <div key="resource" id={self.href} className="item self">
                        <img width="32px" height="32px" className={iconClassName(self)} src={path2Url(self.icon)} alt="" />
                        <div className="name">{self.title}</div>
                    </div>}

                <div className="itemlist" id="itemlistDiv">
                    {links && links.map(l => {
                        return <div key={l.href} id={l.href}
                            className={className(l)}
                            onClick={() => this.select(l)}
                            onDoubleClick={() => this.selectAndActivate(l)}>
                            {l.icon && <img className={iconClassName(l)} src={path2Url(l.icon)} height="20" width="20" />}
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
