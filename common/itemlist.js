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


export class ItemList extends React.Component {

    constructor(props) {
        super(props)

        this.state = { 
            items: this.props.model.items, 
            selectedUrl: this.props.model.selectedUrl
        };
       
        this.props.model.updateListeners.push(() => {
            this.setState({ 
                items: this.props.model.items, 
                selectedItem: this.props.model.selectedItem
            })
        });
        console.log(">>>>>>>>>>>>>>>>>> num listeners now:", this.props.model.updateListeners.length)
    }

    componentDidUpdate = () => {
        // Scroll selected item into view
        if (this.state.selectedItem) {
            let selectedDiv = document.getElementById(this.state.selectedItem.url);
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


    render = () => {
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
                selected={item === this.state.selectedItem}
                onClick={this.props.onClick}
                onDoubleClick={this.props.onDoubleClick} />);
        });

        return  (
            <div id="itemListDiv" style={innerStyle}>
                {content}
            </div>
        )
    }
}
