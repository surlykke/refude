// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';
import {Item} from './item.jsx'

export class ItemList extends React.Component {

    constructor(props) {
        super(props)
        this.state = {items: [], selected: 0};
        this.onUpdated = props.onUpdated
    }

    componentDidUpdate = () => {
        document.getElementById("input").focus();
        // Scroll selected item into view
        let selected = this.state.items[this.state.selected];
        if (selected) {
            let selectedDiv = document.getElementById(selected.Self);
            if (selectedDiv) {
                let listDiv = document.getElementById("itemListDiv");
                let {top: listTop, bottom: listBottom} = listDiv.getBoundingClientRect();
                let {top: selectedTop, bottom: selectedBottom} = selectedDiv.getBoundingClientRect();
                if (selectedTop < listTop) listDiv.scrollTop -= (listTop - selectedTop + 25)
                else if (selectedBottom > listBottom) listDiv.scrollTop += (selectedBottom - listBottom + 10)
            }
        }
        if (this.onUpdated) this.onUpdated()
    };

    componentWillReceiveProps = (props) => {
        let selectedSelf = this.state.items[this.state.selected] && this.state.items[this.state.selected].Self;
        let newSelected = this.state.items.findIndex(i => i.Self === selectedSelf);
        this.setState({items: props.items, selected: newSelected > 0 ? newSelected : 0})
    }

    keyDown = (event) => {
        let {key, ctrlKey, shiftKey, altKey, metaKey} = event;

        if (key === "Tab" && !ctrlKey && shiftKey && !altKey && !metaKey) this.move(false);
        else if (key === "Tab" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true);
        else if (key === "ArrowUp" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(false);
        else if (key === "ArrowDown" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true);
        else if (key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.props.execute(this.state.items[this.state.selected]);
        else if (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey) this.props.execute(this.state.items[this.state.selected]);
        else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.props.onDismiss();
        else {
            return;
        }
        event.preventDefault();
    };

    move = (down) => {
        this.setState({selected: (this.state.selected + this.state.items.length + (down ? 1 : -1)) % this.state.items.length})
    };



    select = (item) => {
        let index = this.state.items.indexOf(item);
        this.setState({selected: index > - 1 ? index : 0});
    };

    render = () => {
        let {select} = this.props;
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
            if (item.Group !== prevGroup) {
                content.push(<div key={item.Group} style={headingStyle}>{item.Group}</div>)
                prevGroup = item.Group
            }
            content.push(<Item key={item.Self}
                               item={item}
                               selected={item === this.state.items[this.state.selected]}
                               select={this.select}
                               execute={this.props.execute}/>)
        });

        return (
            <div onKeyDown={this.keyDown} style={outerStyle}>
                <div style={searchBoxStyle}>
                    <input id="input"
                           style={inputStyle}
                           type="search"
                           onChange={event => this.props.onTermChange(event.target.value)}
                           disabled={this.props.disabled} />
                </div>
                <div id="itemListDiv" style={innerStyle}>
                    {content}
                </div>
            </div>
        )
    }
}
