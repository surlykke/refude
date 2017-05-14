import React from 'react';
import {render} from 'react-dom';
import {nwHide, nwSetup, doHttp} from '../common/utils'
import {MakeServiceProxy} from "../common/service-proxy"
import {SearchBox, Windows} from "./components.jsx"
import {List} from "../common/components"

const windowsProxy = MakeServiceProxy("http://localhost:7938/wm-service/windows", "http://localhost:7938/wm-service/notify")
const appsProxy = MakeServiceProxy("http://localhost:7938/desktop-service/applications", "http://localhost:7938/desktop-service/notify")

const includeWindow = (term, window) => {
	let res = window &&
		   !window.States.includes("_NET_WM_STATE_ABOVE") &&
		   !["Refude Do", "refudeDo"].includes(window.Name) &&
		   window.Name.toUpperCase().includes(term)
	return res
}

const includeApp = (term, app) =>
	app &&
	term !== "" &&
	app.Name.toUpperCase().includes(term)


class Container extends React.Component {

	constructor(props) {
		super(props)
		this.state = {listOfLists: [], windows: [], searchTerm: ""}
	}

	componentDidMount = () => {
		windowsProxy.subscribe(url => { this.update() })
		appsProxy.subscribe(url => { this.update() })
		this.update()
	}

	update = () => {
		if (! this.updatePending) {
			this.updatePending = true
			setTimeout(() => {
				console.log("Updating...")
				let term = this.state.searchTerm.toUpperCase().trim()
				let listOfLists = [
					{
						desc: "Switch to",
					 	items: windowsProxy.resources().filter(win => includeWindow(term, win))
				    },
					{	desc: "Launch",
					 	items: appsProxy.resources().filter(app => includeApp(term, app))
					}
				]
				this.setState({listOfLists: listOfLists.filter(t => t.items.length > 0), windows: listOfLists[0].items})
				this.updatePending = false
			},
			20)
		}
	}

	onTermChange = event => {
		this.setState({searchTerm: event.target.value})
		this.update()
	}

	execute = (item) => {
		doHttp(item.url, "POST").then(response => {
			this.setState({searchTerm: ""})
			this.update()
			nwHide()
		})
	}

	onKeyDown = event => {
		if (["Tab", "ArrowDown", "ArrowUp", "Enter", " "].includes(event.key)) {
			this.list.onKeyDown(event)
		}
		else if ("Escape" === event.key) {
	    	nwHide()
		}
	}

	render = () => {
		let {windows, apps, selected, searchTerm} = this.state
		console.log("listOfLists: ", this.state.listOfLists)
		return (
			<div className="content">
				<div className="topdown" onKeyDown={this.onKeyDown}>
					<SearchBox onTermChange={this.onTermChange} searchTerm={searchTerm}/>
					<List listOfLists={this.state.listOfLists} execute={this.execute} ref={list => this.list = list}/>
				</div>
				<Windows windows={windows} selected={selected}/>
			</div>
		)}
}

nwSetup()

render(
	<Container/>,
	document.getElementById('root')
);
