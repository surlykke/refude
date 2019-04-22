// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';
import {Item} from './item'
import {publish, subscribe} from "./utils";

export class ItemList extends React.Component {

    constructor(props) {
        super(props)
        this.state = {items: [], selectedUrl: null};
    }

    componentDidMount = () => {
        subscribe("moveRequested", this.move);
        subscribe("click", this.setSelected);
        subscribe("doubleclick", item => publish("itemLaunched", item));
    };

    componentDidUpdate = () => {
        document.getElementById("input").focus();
        // Scroll selected item into view
        if (this.state.selectedUrl) {
            let selectedDiv = document.getElementById(this.state.selectedUrl);
            if (selectedDiv) {
                let listDiv = document.getElementById("itemListDiv");
                let {top: listTop, bottom: listBottom} = listDiv.getBoundingClientRect();
                let {top: selectedTop, bottom: selectedBottom} = selectedDiv.getBoundingClientRect();
                if (selectedTop < listTop) listDiv.scrollTop -= (listTop - selectedTop + 25)
                else if (selectedBottom > listBottom) listDiv.scrollTop += (selectedBottom - listBottom + 10)
            }
        }
        publish("componentUpdated");
    };

    componentWillReceiveProps = (props) => {
        if (!(this.state.selectedUrl && props.items.findIndex(i => this.state.selectedUrl === i.url) > -1)) {
            this.setSelected(props.items[0]);
        }
        this.setState({items: props.items});
    };

    keyDown = (event) => {
        let {key, ctrlKey, shiftKey, altKey, metaKey} = event;

        if (key === "Tab" && !ctrlKey && shiftKey && !altKey && !metaKey) this.move(false);
        else if (key === "Tab" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true);
        else if (key === "ArrowUp" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(false);
        else if (key === "ArrowDown" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true);
        else if (key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) publish("itemLaunched", this.getSelected());
        else if (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey) publish("itemLaunched", this.getSelected());
        else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) publish("dismiss");
        else {
            return;
        }
        event.preventDefault();
    };

    move = (down) => {
        if (this.state.selectedUrl) {
            let index = this.state.items.findIndex(i => this.state.selectedUrl === i.url);
            if (index > -1) {
                let numItems = this.state.items.length;
                index = (index + numItems + (down ? 1 : -1)) % numItems;
                this.setSelected(this.state.items[index]);
            }
        }
    };

    setSelected = (item) => {
        this.setState({selectedUrl: item ? item.url : undefined});
        publish("boundsBecame", item ? item.bounds : undefined);
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
            marginTop: "4px",
        };

        let inputStyle = {
            width: "100%",
            height: "36px",
            borderRadius: "5px",
            outlineStyle: "none",
        };

        let innerStyle = {
            marginTop: "8px",
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
            content.push(<Item key={item.url} item={item} selected={item.url === this.state.selectedUrl}/>);
        });

        let onTermChange = (event) => publish("termChanged", event.target.value);

        return (
            <div onKeyDown={this.keyDown} style={outerStyle}>
                <div style={searchBoxStyle}>
                    <input id="input" style={inputStyle} type="search" onChange={onTermChange} disabled={this.props.disabled} autoComplete="off"/>
                </div>
                <div id="itemListDiv" style={innerStyle}>
                    {content}
                </div>
            </div>
        )
    }
}
