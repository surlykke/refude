//Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.


import React from 'react'
import { DoItem } from "./doitem"
import { getUrl} from '../common/monitor';
import { ipcRenderer } from 'electron';
import { keyDownHandler } from './keyhandler';
import './Do.css'
export class ResourceList extends React.Component {
    constructor(props) {
        super(props);
        this.state = { resources: [], term: "", index: 0 }
        ipcRenderer.on("doReset", this.reset)
        ipcRenderer.on("doMove", (evt, up) => up ? this.up() : this.down())

    };

    componentDidMount = () => this.setTerm(this.props.term)

    componentDidUpdate = () => {
        // Scroll selected item into view
        let selectedSelf = this.state.resources[this.state.index] && this.state.resources[this.state.index].Self
        let selectedDiv = selectedSelf && document.getElementById(selectedSelf);
        if (selectedDiv) {
            let listDiv = document.getElementById("itemListDiv");
            if (listDiv) {
                let { top: listTop, bottom: listBottom } = listDiv.getBoundingClientRect();
                let { top: selectedTop, bottom: selectedBottom } = selectedDiv.getBoundingClientRect();
                if (selectedTop < listTop) listDiv.scrollTop -= (listTop - selectedTop + 25)
                else if (selectedBottom > listBottom) listDiv.scrollTop += (selectedBottom - listBottom + 10)
            }
        }

        ipcRenderer.send("doResourceSelected", this.selectedResource()) // TODO Optimize
    };

    updateResourceList = () => getUrl(`/search/desktop\?term=${this.state.term}`, resp => this.setState({ resources: resp.data, index: 0 }))
    selectedResource = () => this.state.resources[this.state.index]
    setTerm = (term) => this.setState({ term: term.toLowerCase() }, this.updateResourceList)
    curRes = () => this.state.resources[this.state.index]
    resLen = () => this.state.resources.length
    index = () => this.state.index
    up = () => this.setState({ index: this.resLen() > 0 ? (this.index() + this.resLen() - 1) % this.resLen() : 0 })
    down = () => this.setState({ index: this.resLen() > 0 ? (this.index() + 1) % this.resLen() : 0 })
    right = () => this.curRes() && this.curRes().Actions && this.props.showRes(this.curRes(), this.state.term)
    post = () => this.curRes() && this.props.post(this.curRes().Self)
    del = () => this.curRes() && this.props.del(this.curRes().Self)
    reset = () => this.setState({ term: "", index: 0 }, this.updateResourceList)
    
    keyHandler = keyDownHandler(this.up, this.down, undefined, this.right, this.post, this.del, this.props.dismiss)

    render = () => <>
        <div key="searchBox" className="searchbox" onKeyDown={this.keyHandler}>
            <input id="input"
                className="searchinput"
                type="search"
                onChange={e => this.setTerm(e.target.value)}
                value={this.state.term}
                autoComplete="off"
                autoFocus />
        </div>
        <div key="itemListDiv" id="itemListDiv" className="itemlist">
            {this.state.resources.map((r, i, resources) =>
                <DoItem key={r.Self} res={r}
                    selected={this.state.index === i}
                    onClick={() => this.setState({ index: i })}
                    onDoubleClick={() => this.setState({ index: i }, this.post)} />)
            }
        </div>
    </>
}
