import React from 'react';
import {render} from 'react-dom';
import {nwHide, nwSetup, doHttp} from '../common/utils'
import {MakeServiceProxy} from "../common/service-proxy"
import {List} from "../common/components"
import {Windows} from "./windows.jsx"

const windows = {
	id: "windows",
	desc: "Switch to",
	proxy: MakeServiceProxy("http://localhost:7938/wm-service", "/windows/")
}

const others = [{
		id: "applications",
		desc: "Launch application",
		proxy: MakeServiceProxy("http://localhost:7938/desktop-service", "/applications/")
	}, {
		id: "poweractions",
		desc: "Leave...",
		proxy: MakeServiceProxy("http://localhost:7938/power-service", "/actions/")
	},
]

const includeWindow = (term, window) => {
	return window &&
		   !(window.States || []).includes("_NET_WM_STATE_ABOVE") &&
		   !["Refude Do", "refudeDo"].includes(window.Name) &&
		   window.Name.toUpperCase().includes(term)
}

const includeOther = (term, item) =>
	item &&
	term !== "" &&
	item.Name.toUpperCase().includes(term)


class Container extends React.Component {

	constructor(props) {
		super(props)
		this.state = {listOfLists: [], windows: [], searchTerm: ""}
		this.allItems = []

		nwSetup((argv) => {
			this.readArgs(argv)
		})
	}

	componentDidMount = () => {
		windows.proxy.subscribe(url => { this.update() })
		others.forEach(other => {
			other.proxy.subscribe(url => { this.update() })
		})
		this.update()
	}

	update = () => {
		if (! this.updatePending) {
			this.updatePending = true
			setTimeout(() => {
				let term = this.state.searchTerm.toUpperCase().trim()
				let listOfLists = [ {
					desc: windows.desc,
				 	items: windows.proxy.resources().filter(win => includeWindow(term, win))
			    }]
				others.forEach(other => {
					listOfLists.push({
						desc: other.desc,
						items: other.proxy.resources().filter(res => includeOther(term, res))
					})
				})
				this.setState({listOfLists: listOfLists.filter(t => t.items.length > 0), windows: listOfLists[0].items})
				this.allItems = []
				listOfLists.forEach(t => {this.allItems.push(...t.items)})
				if (! (this.state.selected && this.allItems.includes(this.state.selected))) {
					this.setState({selected: this.allItems[0]})
				}
				this.updatePending = false
			},
			20)
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

	execute = () => {
		let item = this.state.selected
		if (item) {
			doHttp(item.url, "POST").then(response => {
				this.setState({selected: undefined, searchTerm: ""})
				this.update()
				nwHide()
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
			"Escape" : () => {nwHide()}
		}[key]

		if (op) {
			op()
			event.preventDefault()
		}
	}

	readArgs = (args) => {
		if (args.indexOf("-u") > -1) {
			this.move(true)
		} else if (args.indexOf("-d") > -1) {
			this.move(false)
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
