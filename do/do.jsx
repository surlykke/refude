import React from 'react';
import {render} from 'react-dom';
import {NW, nwHide, nwSetup, doHttp} from '../common/utils'
import {MakeCollection} from "../common/resource-collection"
import {List} from "../common/components"
import {Windows} from "./windows.jsx"

class Container extends React.Component {

	constructor(props) {
		super(props)
		this.state = {
			listOfLists: [],
			windows: [],
			searchTerm: ""
		}

		this.allItems = []

		this.collectionIds = ["windows", "applications", "poweractions"]
		this.collectionDescription = {
			windows: "Switch to",
			applications: "Launch",
			poweractions: "Leave"
		}
		this.collection = {
			windows:MakeCollection("wm-service", "/windows", this.update),
			applications: MakeCollection("desktop-service", "/applications", this.update),
			poweractions: MakeCollection("power-service", "/actions", this.update)
		}

		this.updatePending = false

		nwSetup((argv) => {
			this.readArgs(argv)
		})
	}


	componentDidMount = () => {
		this.readArgs(NW.App.argv)
		this.update()
	}

	updateHelper = () => {
		let listOfLists = []
		this.allItems = []
		let term = this.state.searchTerm.toUpperCase().trim()

		let addMatching = (id) => {
			let items = this.collection[id].filter(res => res.Name.toUpperCase().includes(term))
			if (items.length > 0) {
				listOfLists.push({desc: this.collectionDescription[id], items: items})
				this.allItems.push(...items)
			}
		}

		if (this.onlyShow) {
			if (this.collectionIds.includes(this.onlyShow)) {
				addMatching(this.onlyShow)
			}
		}
		else {
			addMatching("windows")
			if (term !== "") {
				this.collectionIds.slice(1).forEach(id => {
					addMatching(id)
				})
			}
		}
		let windows = this.collection["windows"].filter(res => true)
		this.setState({listOfLists: listOfLists, windows: windows})
		if (! (this.state.selected && this.allItems.includes(this.state.selected))) {
			this.setState({selected: this.allItems[0]})
		}
		this.updatePending = false
	}

	update = () => {
		if (!this.updatePending) {
			this.updatePending = true
			setTimeout(
				this.updateHelper,
				100
			)
		}
	}

	onTermChange = event => {
		this.setState({searchTerm: event.target.value})
		this.update()
	}

	select = (item, execute) => {
		this.setState({selected: item})
		if (execute) this.execute()
	}

	dismiss = () => {
		this.onlyShow = undefined
		this.setState({selected: undefined, searchTerm: ""})
		this.update()
		nwHide()
	}

	execute = () => {
		let item = this.state.selected
		if (item) {
			doHttp(item.url, "POST").then(response => {
				this.dismiss()
			})
		}
	}

	extraClasses = item => {
		let result = ""
		if (item.X !== undefined )  { // So it's a window
			result +=  "window"
			if ((item.States || []).includes("_NET_WM_STATE_HIDDEN")) {
				result += " minimized"
			}
		}

		return result
	}

	move = up => {
		let i = this.allItems.indexOf(this.state.selected)
		i = (i + (up ? -1 : 1) + this.allItems.length) % this.allItems.length
		this.select(this.allItems[i])
	}


	onKeyDown = (event) => {
		let {key, ctrlKey, shiftKey, altKey, metaKey} = event
		let op = {
			Tab: () => {this.move(shiftKey)},
	        ArrowDown : () => {this.move()},
	        ArrowUp :  () => {this.move(true)},
	        Enter : () => {this.select(this.state.selected, true)},
	        " " : () => {this.select(this.state.selected, true)},
			"Escape" : this.dismiss
		}[key]

		if (op) {
			op()
			event.preventDefault()
		}
	}

	readArgs = (args) => {
		console.log("readArgs", args)
		if (args.includes("refude::up")) {
			this.move(true)
		}
		else if (args.includes("refude::down")) {
			this.move(false)
		}
		else {
			let onlyShowArg = args.find(arg => arg.startsWith("refudeOnlyShow::"))
			console.log("onlyShowArg:", onlyShowArg)
			if (onlyShowArg) {
				this.onlyShow = onlyShowArg.slice("refudeOnlyShow::".length)
			}
			this.update()
		}
	}

	render = () => {
		let {windows, apps, selected, searchTerm} = this.state
		return (
			<div className="content">
				<div className="topdown" onKeyDown={this.onKeyDown}>
				    <div className="searchInput" onChange={this.onTermChange} >
				        <input type="search" autoFocus value={searchTerm}/>
				    </div>

					<List listOfLists={this.state.listOfLists}
						  select={this.select}
						  selected={this.state.selected}
						  extraClasses={this.extraClasses}/>
				</div>
				<Windows windows={windows} selected={selected}/>
			</div>
		)}
}

render(
	<Container/>,
	document.getElementById('root')
);
