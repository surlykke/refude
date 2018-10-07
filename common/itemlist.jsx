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
        this.state = {items: []};
    }

    componentDidUpdate = () => {
        document.getElementById("input").focus();
        // Scroll selected item into view
        if (this.state.selected) {
            let selectedDiv = document.getElementById(this.state.selected._self);
            if (selectedDiv) {
                let listDiv = document.getElementById("itemListDiv");
                let {top: listTop, bottom: listBottom} = listDiv.getBoundingClientRect();
                let {top: selectedTop, bottom: selectedBottom} = selectedDiv.getBoundingClientRect();
                if (selectedTop < listTop) listDiv.scrollTop -= (listTop - selectedTop + 25)
                else if (selectedBottom > listBottom) listDiv.scrollTop += (selectedBottom - listBottom + 10)
            }
        }
    };

    componentWillReceiveProps = (props) => {
        this.itemMap = props.items;
        this.filterAndSort();

    };

    filterAndSort = () => {
        let items = [];
        let term = document.getElementById("input").value.toUpperCase();
        for (let [groupName, groupItems] of this.itemMap) {
            groupItems.forEach(item => {
                item.__group = groupName ? groupName : undefined;
                if ((item.__minTermSize || 0) > term.length) {
                    item.__weight = -1;
                } else {
                    item.__weight = item.Name.toUpperCase().indexOf(term);
                    if (item.__weight < 0 && item.Comment) {
                        item.__weight = item.Comment.toUpperCase().indexOf(term);
                        if (item.__weight >= 0) {
                            item.__weight += 100;
                        }
                    }
                }
            });
            let tmp = groupItems.filter(item => item.__weight >= 0).sort((i1, i2) => i1.__weight - i2.__weight);
            items.push(...tmp);
        }

        if (items.length > 0) {
            let prev = items[items.length - 1];
            items.forEach(item => {
                item.__prev = prev;
                prev.__next = item;
                prev = item;
            });
        }

        let selected = undefined;
        if (this.state.selected) {
            selected = items.find(item => item._self === this.state.selected._self);

        }

        if (!selected) {
            selected = items[0];
        }

        this.setState({items: items});
        this.select(selected);
    };


    keyDown = (event) => {
        let {key, ctrlKey, shiftKey, altKey, metaKey} = event;

        if (key === "Tab" && !ctrlKey && shiftKey && !altKey && !metaKey) this.move(false);
        else if (key === "Tab" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true);
        else if (key === "ArrowUp" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(false);
        else if (key === "ArrowDown" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true);
        else if (key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.props.execute(this.state.selected);
        else if (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey) this.props.execute(this.state.selected);
        else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.props.onDismiss();
        else {
            return;
        }
        event.preventDefault();
    };

    move = (down) => {
        if (this.state.selected) {
            let newSelected = down ? this.state.selected.__next : this.state.selected.__prev;
            this.select(newSelected);
        }
    };

    select = (item) => {
        this.setState({selected: item});
        this.props.select(item);
    };

    clear = () => {
        document.getElementById("input").value = "";
    };

    render = () => {
        let {select} = this.props;
        let outerStyle = {
            display: "flex",
            flexFlow: "column",
            height: "100%",
            paddingTop: "0.3em",
            paddingLeft: "0.3em",
        };

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
            if (item.__group !== prevGroup) {
                content.push(<div key={item.__group} style={headingStyle}>{item.__group}</div>)
                prevGroup = item.__group
            }
            content.push(<Item key={item._self}
                               item={item}
                               selected={item === this.state.selected}
                               select={this.select}
                               execute={this.props.execute}/>)
        });

        return (
            <div onKeyDown={this.keyDown} style={outerStyle}>
                <div style={searchBoxStyle}>
                    <input id="input"
                           style={inputStyle}
                           type="search"
                           onChange={this.filterAndSort}
                           disabled={this.props.disabled}
                           autoFocus />
                </div>
                <div id="itemListDiv" style={innerStyle}>
                    {content}
                </div>
            </div>
        )
    }
}
