//Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.


import React from 'react'
import ReactDOM from 'react-dom'
import { DoItem} from "./doitem"
import { getUrl, postUrl, deleteUrl } from '../common/monitor';
import { ipcRenderer} from 'electron';


export class Do extends React.Component {
    constructor(props) {
        super(props);
        this.state = { resources: [], term: "" }
        this.history = []
        this.url = "/search/desktop"
        ipcRenderer.on("doReset", this.reset)
        ipcRenderer.on("doMove", (evt, down) => this.move(down))

    };

    componentDidMount = () => {
        this.updateResourceList()
    }

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
    };

    updateResourceList = () => {
        let separator = this.url.indexOf('?') > -1 ? '&' : '?';
        getUrl(this.url + `${separator}term=${this.state.term}`, resp => {
            this.setState({ resources: resp.data, selected: resp.data[0] && resp.data[0].Self })
        })
    }

    selectedResource = () => {
        let res = this.state.selected && this.state.resources.find(r => r.Self === this.state.selected)
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
        else if (key === "k" && ctrlKey && !shiftKey && !altKey && !metaKey) this.move(false);
        else if (key === "ArrowDown" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true); 
        else if (key === "j" && ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true);
        else if (key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.activate();
        else if (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey) this.activate();
        else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) ipcRenderer.send("doClose");
        else if (key === "Delete" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.delete();
        else if (key === "ArrowRight" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.navigate();
        else if (key === "l" && ctrlKey && !shiftKey && !altKey && !metaKey) this.navigate();
        else if (key === "ArrowLeft" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.navigateBack();
        else if (key === "h" && ctrlKey && !shiftKey && !altKey && !metaKey) this.navigateBack();
        else {
            return;
        }
        event.preventDefault();
    };

    move = (down) => {
        let index = this.state.resources.findIndex(r => r.Self === this.state.selected)
        if (index > -1) {
            index = (index + this.state.resources.length + (down ? 1 : -1)) % this.state.resources.length;
            this.setState({ selected: this.state.resources[index].Self })
        }
    }

    select = (url) => {
        this.setState({ selected: url })
    }

    activate = (url) => {
        url = url || this.state.selected
        if (url) {
            postUrl(url, response => { ipcRenderer.send("doClose") });
        }
    }

    delete = (url) => {
        url = url || this.state.selected;
        url && deleteUrl(url, response => ipcRenderer.send("doClose"));
    }

    navigate = () => {
        let res = this.selectedResource()
        if (res && res.OtherActions) {
            this.history.unshift({url: this.url, term: this.state.term, selected: this.state.selected})
            this.url = res.OtherActions;
            this.setState({term: ""}, this.updateResourceList)
        }
    }

    navigateBack = () => {
        if (this.history.length > 0) {
            let history = this.history.shift()
            this.url = history.url;
            this.setState({term: history.term, selected: history.selected}, this.updateResourceList)
        }
    }

    reset = () => {
        this.url = "/search/desktop"
        this.history = []
        this.setState({ term: "" }, this.updateResourceList)
    };

    render = () => {

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

        let itemListStyle = {
            flexGrow: "1",
            overflowY: "scroll"
        };

        ipcRenderer.send("doResourceSelected", this.selectedResource()) // TODO Optimize

        return <>
            <div key="searchBox" style={searchBoxStyle} onKeyDown={this.keyDown}>
                <input id="input"
                    style={inputStyle}
                        type="search"
                        onChange={e => this.setTerm(e.target.value)}
                        value={this.state.term}
                        autoComplete="off"
                        autoFocus />
            </div>
            <div key="itemListDiv" id="itemListDiv" style={itemListStyle}>
                {this.state.resources.map((r,i,resources) => 
                    <DoItem key={r.Self} res={r} selected={this.state.selected === r.Self} 
                            onClick={() => this.select(r)} onDoubleClick={() => this.activate(r.Self)}/>)
                }
            </div>
        </>
    }
}

ReactDOM.render(<Do/>,document.getElementById('do'))