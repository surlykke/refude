// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';
import ReactDom from 'react-dom'
import {Item} from './item.jsx'
import {SearchBox} from "./searchbox";

export let linkItems = (items) => {
    if (items.length > 0) {
        let prev = items[items.length - 1];
        items.forEach(item => {
            item.__prev = prev;
            prev.__next = item;
            prev = item;
        });
    }
};

export class ItemList extends React.Component {

    constructor(props) {
        super(props)
        this.state = {};
        this.searchBox = React.createRef();
    }

    componentDidMount = () => {
        console.log("componentDidMount");
    };

    componentDidUpdate = () => {
        console.log("componentDidUpdate");
    };

    componentWillReceiveProps = (newProps) => {
        let newSelected = undefined;
        if (this.state.selected) {
            console.log("looking for selected");
            newSelected = newProps.items.find(item => item._self === this.state.selected._self);

        }
        if (!newSelected) {
            newSelected = newProps.items[0];
        }

        this.setState({selected: newSelected});
    };


    /*if (this.props.selectedSelf) {
        let selectedDiv = document.getElementById(this.props.selectedSelf)
        if (selectedDiv) {
            let listDiv = document.getElementById("itemListDiv")
            let {top: listTop, bottom: listBottom} = listDiv.getBoundingClientRect()
            let {top: selectedTop, bottom: selectedBottom} = selectedDiv.getBoundingClientRect()
            if (selectedTop < listTop) listDiv.scrollTop -=  (listTop - selectedTop + 25)
            else if (selectedBottom > listBottom) listDiv.scrollTop += (selectedBottom - listBottom + 10)
        }
    }*/

    keyDown = (event) => {
        console.log("keydown:", event);
        let {key, ctrlKey, shiftKey, altKey, metaKey} = event;
        if (key === "Tab" && !ctrlKey && shiftKey && !altKey && !metaKey) this.move(false);
        else if (key === "Tab" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true);
        else if (key === "ArrowUp" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(false);
        else if (key === "ArrowDown" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true);
        else if (key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.execute(this.state.selected);
        else if (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey) this.execute(this.state.selected);
        else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.dismiss();
        else {
            return;
        }

        event.stopPropagation();
    };

    move = (down) => {
        console.log("move, this.state.selected:", this.state.selected, "down:", down);
        if (this.state.selected) {
            let newSelected = down ? this.state.selected.__next : this.state.selected.__prev;
            console.log("newSelected:", newSelected);
            this.setState({selected: newSelected});
        }
    };


    select = (item) => {
        this.setState({selected: item});
        this.props.select(item);
    };

    execute = (item) => {
        this.props.execute(item);
        this.dismiss();
    };

    dismiss = () => {
        console.log(this.searchBox.current.clear)
        this.searchBox.current.clear();
        //this.props.onTermChange("");
        this.props.onDismiss();
    };


    render = () => {
        let {items, onTermChange, select} = this.props
        let style = {
            overflow: "auto",
        }
        Object.assign(style, this.props.style)

        let headingStyle = {
            fontSize: "0.9em",
            color: "gray",
            fontStyle: "italic",
            marginTop: "5px",
            marginBottom: "3px",
        };

        let prevGroup
        let content = []
        items.forEach(item => {
            if (item.__group !== prevGroup) {
                content.push(<div style={headingStyle}>{item.__group}</div>)
                prevGroup = item.__group
            }
            content.push(<Item key={item._self}
                               item={item}
                               selected={item === this.state.selected}
                               select={this.select}
                               execute={this.execute}/>)
        })
        return (
            <div onKeyDown={this.keyDown}>
                <SearchBox onChange={onTermChange} ref={this.searchBox}/>
                <div id="itemListDiv" style={style}>
                    {content}
                </div>
            </div>
        )
    }
}
