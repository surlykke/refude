import React from 'react';
import {render} from 'react-dom';
import {nwHide, nwSetup, doHttp} from '../common/utils'
import {MakeServiceProxy} from "../common/service-proxy"
import {SearchBox, CommandList,Windows} from "./components.jsx"

const windowsProxy = MakeServiceProxy("http://localhost:7938/wm-service/windows", "http://localhost:7938/wm-service/notify")
const appsProxy = MakeServiceProxy("http://localhost:7938/desktop-service/applications", "http://localhost:7938/desktop-service/notify")

const includeWindow = (term, window) =>
	window &&
	!window.States.includes("_NET_WM_STATE_ABOVE") &&
	!["Refude Do", "refudeDo"].includes(window.Name) &&
	window.Name.toUpperCase().includes(term)

const includeApp = (term, app) =>
	app &&
	term !== "" &&
	app.Name.toUpperCase().includes(term)

const updateState = (currentState, newSearchTerm) => {
	if (newSearchTerm === undefined) {
		newSearchTerm = currentState.searchTerm
	}
	let term = newSearchTerm.toUpperCase().trim()
	let newState = {
		windows: windowsProxy.index().map(url => windowsProxy.get(url)).filter(win => includeWindow(term, win)),
		apps: term === "" ? [] : appsProxy.index().map(url => appsProxy.get(url)).filter(app => includeApp(term, app)),
		searchTerm: newSearchTerm
	}

	if (!(newState.windows.includes(currentState.selected) || newState.apps.includes(currentState.selected))) {
		newState.selected = newState.windows[0] || newState.apps[0]
	}

	return newState
}

class Container extends React.Component {

	constructor(props) {
		super(props)
		this.state = {windows: [], apps: [], searchTerm: ""}
	}

	componentDidMount = () => {
		windowsProxy.subscribe(url => {
			if (url === windowsProxy.indexUrl || windowsProxy.index().includes(url)) {
				this.setState(updateState(this.state))
			}
		})

		appsProxy.subscribe(url => {
			if (this.state.searchTerm !== "" && (url === appsProxy.indexUrl || appsProxy.index().includes(url))) {
				this.setState(updateState(this.state))
			}
		})

		this.setState(updateState(this.state))

		console.log("this.refs: ", this.refs)
	}


	onTermChange = event => {
		this.setState(updateState(this.state, event.target.value))
	}

	select = (res, exec) => {
		this.setState({selected: res})
		if (exec) {
			this.execute()
		}
	}

	move = up => {
		let allRes = this.state.windows.concat(this.state.apps)
		if (allRes.length > 0) {
			let index = allRes.findIndex(res => res.url === this.state.selected.url)
			let newSelected = allRes[(index + (up ? -1 : 1) + allRes.length) % allRes.length]
			this.setState({selected: newSelected})
		}
	}

	execute = () => {
		if (this.state.selected) {
			doHttp(this.state.selected.url, "POST").then(response => {
				this.setState({searchTerm: ""})
				nwHide()
			})
		}
	}

	onKeyDown = event => {
		let {key, ctrlKey, shiftKey, altKey, metaKey} = event.nativeEvent
		let op = {
			Tab: () => {this.move(shiftKey)},
	        ArrowDown : () => {this.move()},
	        ArrowUp :  () => {this.move(true)},
	        Enter : () => {this.execute()},
	        " " : () => {this.execute()},
	        Escape : () => {nwHide()}
		}[key]

		if (op) op()
	}

	render = () => {
		let {windows, apps, selected, searchTerm} = this.state
		return (
			<div className="content">
				<div className="left" onKeyDown={this.onKeyDown}>
					<SearchBox onTermChange={this.onTermChange} searchTerm={searchTerm}/>
					<CommandList windows={windows} apps={apps} selected={selected} select={this.select}/>
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
