// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';
import { Item } from './item'
import { publish, subscribe } from "./utils";
import { timingSafeEqual } from 'crypto';

let match = (item, term) => {
    return     term === "" ? item.showInitially : item.name.toLocaleLowerCase().indexOf(term) > -1 || item.comment.toLocaleLowerCase().indexOf(term) > -1
}

let rank = (item , lowercaseTerm) => {
    let baserank = item.rank || 0
    let tmp;
    if ((tmp = item.name.toLowerCase().indexOf(lowercaseTerm)) > -1) {
        return baserank -tmp;
    } else if (item.comment && (tmp = item.comment.toLowerCase().indexOf(lowercaseTerm)) > -1) {
        return baserank -tmp - 100;
    } else {
        return 1;
    }
};



export class ItemList extends React.Component {

    constructor(props) {
        super(props)
        this.state = { items: [], selectedUrl: null };
    }

    componentDidMount = () => {
        subscribe("reset", this.reset)
        subscribe("moveRequested", this.move);
    };

    componentDidUpdate = () => {
        document.getElementById("input").focus();
        // Scroll selected item into view
        if (this.state.selectedUrl) {
            let selectedDiv = document.getElementById(this.state.selectedUrl);
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

    componentWillReceiveProps = (props) => {
        document.getElementById("input").value = ""
        let items = this.filter(props.items)
        this.setState({ items: items});
        this.ensureSelection(items)
    };

    filter = items => {
        let term = document.getElementById("input").value.toLocaleLowerCase()
        return items.filter(item => match(item, term)).sort((i1, i2) => rank(i2, term) - rank(i1, term))
    }

    onTermChange = () => {
        let items = this.filter(this.props.items)
        this.setState({ items: items })
        this.ensureSelection(items)   
    }

    ensureSelection = items => {
        if (!(this.state.selectedUrl && items.findIndex(i => this.state.selectedUrl === i.url) > -1)) {
            this.setSelected(items[0])
        }
    }

    reset = () => {
        document.getElementById("input").value = ""
        this.setState({items: [], selectedUrl: undefined})
    }

    keyDown = (event) => {
        let { key, ctrlKey, shiftKey, altKey, metaKey } = event;

        if (key === "Tab" && !ctrlKey && shiftKey && !altKey && !metaKey) this.move(false);
        else if (key === "Tab" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true);
        else if (key === "ArrowUp" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(false);
        else if (key === "ArrowDown" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true);
        else if (key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.props.activate(this.getSelected());
        else if (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey) this.props.activate(this.getSelected());
        else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.props.dismiss();
        else {
            return;
        }
        event.preventDefault();
    };

    move = (down) => {
        let index = this.state.items.findIndex(i => this.state.selectedUrl === i.url);
        if (index > -1) {
            let numItems = this.state.items.length;
            index = (index + numItems + (down ? 1 : -1)) % numItems;
        } else {
            index = 0
        }
        this.setSelected(this.state.items[index]);
    };

    setSelected = (item) => {
        let selectedUrl = item ? item.url : undefined
        this.setState({ selectedUrl: selectedUrl });
        this.props.select(selectedUrl)
    };

    getSelected = () => {
        return this.state.items.find(i => i.url === this.state.selectedUrl);
    }

    render = () => {
        let outerStyle = Object.assign({
            display: "flex",
            flexFlow: "column",
            paddingTop: "0.3em",
            paddingLeft: "0.3em",
        }, this.props.style);

        let searchBoxStyle = {
            boxSizing: "border-box",
            paddingRight: "5px",
            width: "calc(100% - 16px)",
            marginTop: "8px"
        };

        let inputStyle = {
            width: "100%",
            height: "36px",
            borderRadius: "5px",
            outlineStyle: "none",
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

        let prevGroup;
        let content = [];
        this.state.items.forEach(item => {
            if (item.group !== prevGroup) {
                content.push(<div key={item.group} style={headingStyle}>{item.group}</div>);
                prevGroup = item.group;
            }
            content.push(<Item key={item.url}
                item={item}
                selected={item.url === this.state.selectedUrl}
                onClick={this.setSelected}
                onDoubleClick={this.props.activate} />);
        });

        return (
            <div onKeyDown={this.keyDown} style={outerStyle}>
                <div style={searchBoxStyle}>
                    <input id="input" value={this.state.term} style={inputStyle} type="search" onChange={this.onTermChange} disabled={this.props.disabled} autoComplete="off" />
                </div>
                <div id="itemListDiv" style={innerStyle}>
                    {content}
                </div>
            </div>
        )
    }
}
