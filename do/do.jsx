// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import {render} from 'react-dom'
import {NW, devtools, nwHide, nwSetup, doGet, doPost, adjustIconUrl} from '../common/utils'
import {ItemList} from "../common/itemlist"
import {SearchBox} from "../common/searchbox"
import {Windows} from "./windows.jsx"

class Container extends React.Component {

	constructor(props) {
		super(props);
		this.state = { items: [], windows: [], searchTerm: "" };
        this.resources = {
		    "wm-service": [],
            "desktop-service": [],
            "power-service": []
        };

		nwSetup((argv) => {
			this.readArgs(argv);
		})
	}

	fetchResources = (searchTerm) => {
		console.log("Fetching for", searchTerm);
	    let query = {q: searchTerm || ""};
	    this.resources["wm-service"] = [];
	    this.resources["desktop-services"] = [];
	    this.resources["power-service"] = [];

	    let showWin = w => !(w.States && w.States.includes("_NET_WM_STATE_ABOVE") || ["Refude Do", "refudeDo"].includes(w.Name));

	    doGet("wm-service", "/search", query).then(resources => {
	        this.resources["wm-service"] = resources.filter(showWin);
	        this.resources["wm-service"].forEach(adjustIconUrl)
	        this.updateItems();
	    });

        if (searchTerm) {
            doGet("desktop-service", "/search", query).then(resources => {
                this.resources["desktop-service"] = resources;
                this.resources["desktop-service"].forEach(adjustIconUrl)
                this.updateItems();
            });
            doGet("power-service", "/search", query).then(resources => {
                this.resources["power-service"] = resources;
                this.resources["power-service"].forEach(adjustIconUrl)
                this.updateItems();
            });
        }
	};

	updateItems = () => {
        let addAll = (dst, src, group) => {
			src.forEach(item => {
				item.group = group
				dst.push(item)
			})
		}
		console.log("Into updateItems, resources:", this.resources);
	    let items = [];
	    addAll(items, this.resources["wm-service"], "Open windows");
        addAll(items, this.resources["desktop-service"], "Applications");
        addAll(items, this.resources["power-service"], "Leave");

        if (items.length > 0) {
            if (! items.find(item => item.Self === this.state.selected)) {
                this.setState({selected: items[0].Self})
            }
        } else {
            this.setState({selected: undefined})
        }

        this.setState({items: items, windows: this.resources["wm-service"]})
    }

	componentDidMount = () => {
		this.readArgs(NW.App.argv)
//		this.fetchResources()
	};

	onTermChange = (searchTerm) => {
		this.setState({searchTerm: searchTerm})
		this.fetchResources(searchTerm);
	};

	onKeyDown = (event) => {
		let {key, ctrlKey, shiftKey, altKey, metaKey} = event
		if      (key === "Tab" && !ctrlKey &&  shiftKey && !altKey && !metaKey) this.move(false)
		else if (key === "Tab" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true)
		else if (key === "ArrowUp" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(false)
		else if (key === "ArrowDown" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true)
		else if (key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.execute(this.state.selected)
		else if (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey) this.execute(this.state.selected)
		else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.dismiss()
		else if (key === "Alt" && !ctrlKey && !shiftKey && altKey && !metaKey) this.collected = 0
		else if ("0" <= key && key <= "9" && !ctrlKey && !shiftKey && altKey && !metaKey && this.collected !== undefined) {
			this.collected = 10*this.collected + key - "0"
		}
		else {
			return
		}

		event.stopPropagation();
	}

	move = (down) => {
		let index = this.state.items.findIndex(item => item.Self === this.state.selected)
		if (index > -1) {
			index = (index + this.state.items.length + (down ? 1 : -1)) % this.state.items.length
			this.setState({selected: this.state.items[index].Self})
		}
	}

	onKeyUp = (event) => {
		if ("Alt" === event.key && this.collected !== undefined) {
			if (this.state.windows[this.collected - 1]) this.execute(this.state.windows[this.collected - 1].Self)
			this.collected = undefined
		}
	}

	select = (self) => {
		this.setState({selected: self})
	}

	execute = (self) => {
		if (self) {
			this.select(self)
			doPost(...self.split(":")).then(response => {this.dismiss()})
		}
	}

	dismiss = () => {
		console.log("dismiss")
		this.onlyShow = undefined
		this.onTermChange("")
		nwHide()
	}

	readArgs = (args) => {
		this.onlyShow = undefined
		if (args.includes("refude::up")) {
			this.move(false)
		}
		else if (args.includes("refude::down")) {
			this.move(true)
		}
		else {
			let onlyShowArg = args.find(arg => arg.startsWith("refude::onlyShow::"))
			if (onlyShowArg) {
				this.onlyShow = onlyShowArg.slice("refude::onlyShow::".length)
			}
			this.fetchResources();
		}
	}

	render = () => {
		let {windows, apps, selected, searchTerm} = this.state
		let setList = list => {this.list = list}

		let contentStyle = {
			position: "relative",
			display: "flex",
			boxSizing: "border-box",
			width: "calc(100% - 1px)",
			height: "calc(100% - 1px)",
			padding: "4px",
		}

		let leftColumnStyle = {
			position: "relative",
			width: "260px",
			height: "100%",
			display: "flex",
			flexDirection: "column",
			margin: "0px"
		}

		let searchBoxStyle = {
			marginBottom: "5px",
		}

		let itemListStyle = {
			flex: "1",
		}

		let windowsStyle = {
			margin: "0px",
			marginLeft: "8px",
			flex: "1",
		}

		return (
			<div style={contentStyle} onKeyDown={this.onKeyDown} onKeyUp={this.onKeyUp}>
				<div style={leftColumnStyle}>
					<SearchBox style={searchBoxStyle} onChange={evt => this.onTermChange(evt.target.value)}  searchTerm={this.state.searchTerm}/>
					<ItemList style={itemListStyle} items={this.state.items} selectedSelf={this.state.selected} select={this.select} execute={this.execute}/>
				</div>
				<Windows style={windowsStyle} windows={this.state.windows} selectedSelf={this.state.selected}/>
			</div>
		)
	}
}

render(
	<Container/>,
	document.getElementById('root')
);
