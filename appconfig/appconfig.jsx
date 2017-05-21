/*
* Copyright (c) 2015, 2016 Christian Surlykke
*
* This file is part of the refude project.
* It is distributed under the GPL v2 license.
* Please refer to the LICENSE file for a copy of the license.
*/
import React from 'react';
import {render} from 'react-dom';
import {doHttp} from '../common/utils'
import {MakeServiceProxy} from "../common/service-proxy"
import {MimetypeList} from "./mimetypelist.jsx"

const mimetypesProxy = MakeServiceProxy("http://localhost:7938/desktop-service/mimetypes", "http://localhost:7938/desktop-service/notify")

const includeMimetype = (term, mimetype) => mimetype && mimetype.Comment.toUpperCase().includes(term)

const updateState = (currentState, newSearchTerm) => {
	if (newSearchTerm === undefined) {
		newSearchTerm = currentState.searchTerm
	}
	let term = newSearchTerm.toUpperCase().trim()
	let newState = {
		mimetypes: mimetypesProxy.index().map(url => mimetypesProxy.get(url)).filter(mimetype => includeMimetype(term, mimetype)),
		searchTerm: newSearchTerm
	}

	if (! newState.mimetypes.includes(currentState.selected)) {
		newState.selected = newState.mimetypes[0]
	}

	return newState
}

class AppConfig extends React.Component {

	constructor(props) {
		super(props)
		this.state = {mimetypes: [], searchTerm: ""}
	}

	componentDidMount = () => {
		this.subscribe()
		this.setState(updateState(this.state))
	}

	subscribe = () => {
		let updateScheduled = false
		mimetypesProxy.subscribe(url => {
			if (!updateScheduled) {
				updateScheduled = true
				setTimeout(() => {
					this.setState(updateState(this.state))
					updateScheduled = false
				}, 20)
			}
		})
	}

	onTermChange = event => {
		this.setState(updateState(this.state, event.target.value))
	}

	select = (res, exec) => {
		this.state.selected = res
		if (exec) {
			this.execute()
		}
	}

	move = up => {
		if (this.state.mimetypes.length === 0) return

		let index = this.state.mimetypes.findIndex(mimetype => mimetype.url === this.state.selected.url)
		let newSelected = this.state.mimetypes[(index + (up ? -1 : 1) + this.state.mimetypes.length) % this.state.mimetypes.length]
		this.setState({selected: newSelected})
	}

	execute = () => {
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
		let {mimetypes, selected, searchTerm} = this.state
		return (
			<div className="content">
				<div className="topdown" onKeyDown={this.onKeyDown}>
					<MimetypeList mimetypes={mimetypes} selected={selected} select={this.select}/>
				</div>
			</div>
		)
	}
}

render(
	<AppConfig/>,
	document.getElementById('root')
);
