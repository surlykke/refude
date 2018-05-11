// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import {render} from 'react-dom'
import {NW, devtools, nwHide, nwSetup, doGet, doPost} from '../common/utils'
import {ItemList} from "../common/itemlist"
import {SearchBox} from "../common/searchbox"

class Container extends React.Component {

	constructor(props) {
		super(props);
		this.state = { items: [], searchTerm: "" };
        this.resources = {
		    "wm-service": [],
            "desktop-service": [],
            "power-service": []
        };

		nwSetup((argv) => {
			this.readArgs(argv);
		})
	}

	fetchResources = (term) => {
		console.log("Fetching for", term);
	    term = term || "";
	    this.resources["wm-service"] = [];
	    this.resources["desktop-service"] = [];
	    this.resources["power-service"] = [];

		let winQuery = {
			type: "application/vnd.org.refude.action+json",
			q: `r.Name ~i '${term}' and r.Name neq 'Refude Do' and r.Name neq 'refudeDo'`
		};
		console.log("winQuery:", winQuery)
	    doGet("wm-service", "/search", winQuery).then(resources => {
	        this.resources["wm-service"] = resources;
	        this.updateItems();
	    }, error => console.log(error));

        if (term && term.length > 0) {
        	let appQuery = {
                type: "application/vnd.org.refude.action+json",
                q: `r.Name ~i '${term}'`
            };
            doGet("desktop-service", "/search", appQuery).then(resources => {
                this.resources["desktop-service"] = resources;
                this.updateItems();
            });

			let powerQuery  = {
				type: 'application/vnd.org.refude.action+json',
				q: `r.Name ~i '${term}'`
			}
            doGet("power-service", "/search", powerQuery).then(resources => {
                this.resources["power-service"] = resources;
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
            if (! items.find(item => item._self === this.state.selected)) {
                this.setState({selected: items[0]._self})
            }
        } else {
            this.setState({selected: undefined})
        }

        this.setState({items: items})
    }

	componentDidMount = () => {
		this.readArgs(NW.App.argv)
		this.fetchResources("")
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
		else {
			return
		}

		event.stopPropagation();
	}

	move = (down) => {
		let index = this.state.items.findIndex(item => item._self === this.state.selected)
		if (index > -1) {
			index = (index + this.state.items.length + (down ? 1 : -1)) % this.state.items.length
			this.setState({selected: this.state.items[index]._self})
		}
	}

	select = (self) => {
		this.setState({selected: self})
	}

	execute = (self) => {
		console.log("Self: ", self)
		if (self) {
			let item = this.state.items.find(i => self === i._self)
			this.select(self)
			doPost(item).then(response => {this.dismiss()})
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
        this.fetchResources("");
		if (args.includes("up")) {
			this.move(false)
		}
		else  {
			this.move(true)
		}
	}

	render = () => {
		let {apps, selected, searchTerm} = this.state
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
			width: "100%",
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


		return (
			<div style={contentStyle} onKeyDown={this.onKeyDown} onKeyUp={this.onKeyUp}>
				<div style={leftColumnStyle}>
					<SearchBox style={searchBoxStyle} onChange={evt => this.onTermChange(evt.target.value)}  searchTerm={this.state.searchTerm}/>
					<ItemList style={itemListStyle} items={this.state.items} selectedSelf={this.state.selected} select={this.select} execute={this.execute}/>
				</div>
			</div>
		)
	}
}

render(
	<Container/>,
	document.getElementById('root')
);
