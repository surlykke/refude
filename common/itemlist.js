// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';
import { Item } from './item'
import { publish, filterAndSort  } from "./utils";


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

export let makeModelAndController = () => {
    let term = "";

    let model = {
        updateListeners: [],
        selectedItem: null, 
        items: [],
        selectedIndex: () => model.selectedItem ? model.items.findIndex(i => i.url === model.selectedItem.url) : -1
    }
 
    let notifyListeners = () => {
        model.updateListeners.forEach(l => l());
    }    

    let controller = {
        itemMap: new Map(),
        setTerm: (t) => {
            term = t.toLowerCase();
            controller.update()
        },
        move: (down) => {
            let index = model.selectedIndex()
            if (index > -1) {
                let numItems = model.items.length;
                index = (index + numItems + (down ? 1 : -1)) % numItems;
                model.selectedItem = model.items[index];
                notifyListeners();
            }
        },
        select: (item) => {
            if (model.items.findIndex(i => i.url === item.url) > - 1) {
                model.selectedItem = item;
                notifyListeners();
            } 
        },
        update: () => { 
            model.items = [];
            controller.itemMap.forEach((itemList, k) => {
                model.items.push(...filterAndSort(itemList, term));
            });
            let index = model.selectedIndex();
            if (index > -1) {
                model.selectedItem = model.items[index]
            } else {
                model.selectedItem = model.items[0]
            }
        
            notifyListeners();
        },
        clear: (keys) => {
           keys.forEach(k => controller.itemMap.set(k, []));
           controller.update();
        }

    }

    return [model, controller];
}