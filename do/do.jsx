// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import {render} from 'react-dom'
import {NW, devtools, nwHide, nwSetup, doHttp} from '../common/utils'
import {MakeCollection} from "../common/resources"
import {ItemList} from "../common/itemlist"
import {SearchBox} from "../common/searchbox"
import {Windows} from "./windows.jsx"

class Container extends React.Component {

	constructor(props) {
		super(props)
		this.state = { items: [], windows: [], searchTerm: "" }

		let match = (item, term) => item.Name.toUpperCase().includes(term)

		this.collections = {
			windows: MakeCollection("wm-service", "/windows", this.scheduleUpdate, (win, term) =>
				!(win.States && win.States.includes("_NET_WM_STATE_ABOVE") || ["Refude Do", "refudeDo"].includes(win.Name)) &&
				match(win, term)
			),
			applications: MakeCollection("desktop-service", "/applications", this.scheduleUpdate, (app, term) =>
				!app.NoDisplay && match(app, term)
			),
			poweractions: MakeCollection("power-service", "/actions", this.scheduleUpdate, (app, term) => match(app, term))
		}

		this.collectionHeadings = {
			windows: "Switch to",
			applications: "Launch",
			poweractions: "Leave",
		}

		nwSetup((argv) => {
			this.readArgs(argv);
		})
	}

	componentDidMount = () => {
		this.readArgs(NW.App.argv)
		this.scheduleUpdate()
	}

	update = () => {
		let items = []
		let windows = []
		if (this.onlyShow) {
			items = this.collections[this.onlyShow].filtered
		}
		else {
			windows.push(...this.collections["windows"].filtered)
			items.push(...this.collections["windows"].filtered)
			if (this.state.searchTerm.trim() !== "") {
				items.push(...this.collections["applications"].filtered)
				items.push(...this.collections["poweractions"].filtered)
			}
		}

		this.setState({items: items, windows: windows})
		items.includes(this.state.selected) || this.setState({selected: items[0]})
		this.updatePending = false
	}

	scheduleUpdate = () => {
		if (!this.updatePending) {
			this.updatePending = true
			setTimeout( this.update, 100 )
		}
	}

	onTermChange = (event) => {
		console.log("onTermChange:", event)
		this.collections["windows"].setterm(event.target.value)
		this.collections["applications"].setterm(event.target.value)
		this.collections["poweractions"].setterm(event.target.value)
		this.setState({searchTerm: event.target.value})
		this.scheduleUpdate()
	}

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
		event.stopPropagation
	}

	move = (down) => {
		let index = this.state.items.indexOf(this.state.selected)
		if (index > -1) {
			index = (index + this.state.items.length + (down ? 1 : -1)) % this.state.items.length
		}
		this.setState({selected: this.state.items[index]})
	}

	onKeyUp = (event) => {
		if ("Alt" === event.key && this.collected !== undefined) {
			if (this.state.windows[this.collected - 1]) this.execute(this.state.windows[this.collected - 1])
			this.collected = undefined
		}
	}

	select = (item) => {
		this.setState({selected: item})
	}

	execute = (item) => {
		this.select(item)
		doHttp(item.url, "POST").then(response => {this.dismiss()})
	}

	dismiss = () => {
		this.onlyShow = undefined
		this.setState({searchTerm: ""})
		this.scheduleUpdate()
		nwHide()
	}

	readArgs = (args) => {
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
			this.scheduleUpdate()
		}
	}

	render = () => {
		let {windows, apps, selected, searchTerm} = this.state
		let setList = list => {this.list = list}

		let contentStyle = {
			position: "relative",
			display: "flex",
			boxSizing: "border-box",
			width: "calc(100% - 16px)",
			height: "calc(100% - 16px)",
			margin: "8px",
		}

		let leftColumnStyle = {
			position: "relative",
			width: "260px",
			height: "100%",
			display: "flex",
			flexDirection: "column",
		}

		let searchBoxStyle = {
			marginBottom: "5px",
		}

		let itemListStyle = {
			flex: "1",
		}

		let windowsStyle = {
			marginLeft: "8px",
			flex: "1",
		}

		return (
			<div style={contentStyle} onKeyDown={this.onKeyDown} onKeyUp={this.onKeyUp}>
				<div style={leftColumnStyle}>
					<SearchBox style={searchBoxStyle} onChange={this.onTermChange}  searchTerm={this.state.searchTerm}/>
					<ItemList style={itemListStyle} items={this.state.items}
						      selected={this.state.selected} select={this.select} execute={this.execute}/>
				</div>
				<Windows style={windowsStyle} windows={this.state.windows} selected={this.state.selected}/>
			</div>
		)
	}
}

render(
	<Container/>,
	document.getElementById('root')
);
