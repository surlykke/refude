import React from 'react';
import {render} from 'react-dom';
import {doHttp, iconServiceUrl} from '../common/utils'
import {MakeCollection} from '../common/resources'
import {ItemList} from "../common/itemlist"
import {Item} from "../common/item"
import {SearchBox} from "../common/searchbox"

let gui = window.require('nw.gui')
let appArgument = gui.App.argv[0]
let mimetypeId = gui.App.argv[1]

class AppChooser extends React.Component {
	constructor(props) {
		super(props)
		this.mimetypeIds = []
		this.mimetypes = new Map()
		this.state = {
			iconUrl: iconServiceUrl(["unknown"], 32),
			comment: mimetypeId,
			apps: [],
			searchTerm: "",
		}
	}

	componentDidMount() {
		this.apps = MakeCollection("desktop-service", "/applications	", this.scheduleUpdate)
		document.body.addEventListener("keydown", this.onKeyDown)
		this.fetch(mimetypeId)
	}

	fetch = (id) => {
		let url = "http://localhost:7938/desktop-service/mimetypes/" + id
		doHttp(url).then(mimetype => {
			if (! this.mimetypeIds.includes(id)) {
				this.mimetypeIds.push(id)
				mimetype.url = url
				mimetype.IconUrl = iconServiceUrl([mimetype.IconName, mimetype.GenericIcon])
				this.mimetypes[id] = mimetype
				mimetype.SubClassOf.forEach(subId => { this.fetch(subId)})
				if (id === mimetypeId) {
					this.setState({iconUrl: mimetype.IconUrl, comment: mimetype.Comment})
				}
				this.update()
			}
		})
	}

	// We get a lot of events from apps, so we collect to, at most, one update pr 20 ms
	scheduleUpdate = () => {
		if (! this.updatePending) {
			this.updatePending = true
			setTimeout(this.update, 20)
		}
	}

	update = () => {
		let term = this.state.searchTerm.toUpperCase().trim()
		let apps = this.apps.filter(app => app.Actions["_default"]["Exec"].match(/%f|%F|%u|%U/))
		                    .filter(app => app.Name.toUpperCase().includes(term))

		apps.forEach(app => {
			let mimetypeId = this.mimetypeIds.find(id => app.Mimetypes.includes(id))
			if (mimetypeId) {
				app.group = "Applications that handle: '" + this.mimetypes[mimetypeId].Comment + "'"
				app.order = this.mimetypeIds.indexOf(mimetypeId)
			} else {
				app.group = "Other applications"
				app.order = this.mimetypeIds.length
			}
			console.log(app.Name: ": ", app.order, app.group)
		})
		apps.sort((app1, app2) => app1.order !== app2.order ? app1.order - app2.order : app1.Name.localeCompare(app2.Name))
		if (!(this.state.selected && apps.includes(this.state.selected))) {
			this.setState({selected: apps[0]})
		}
		this.setState({apps: apps})
		this.updatePending = false
	}

	select = (app) => {
		this.setState({selected: app})
	}

	execute = (app) => {
		this.select(app)
		if (app) {
			if (this.state.remember) {
				let mimetypeUrl = "http://localhost:7938/desktop-service/mimetypes/" + mimetypeId
				doHttp(mimetypeUrl, "POST", {DefaultApplication: app.Id})
			}

			doHttp(app.url, "POST", {Arguments: [appArgument]}).then(response => {gui.App.quit()})
		}
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
		else return;
		event.stopPropagation
	}

	move = (down) => {
		let {apps, selected} = this.state
		let index = apps.indexOf(selected)
		if (index > -1) {
			index = (index + apps.length + (down ? 1 : -1)) % apps.length
		}
		this.setState({selected: apps[index]})
	}

	onTermChange = (event) => {
		console.log("onTermChange:", event)
		this.setState({searchTerm: event.target.value})
		this.scheduleUpdate()
	}

	render = () => {
		let {iconUrl, comment, apps, selected} = this.state
		let contentStyle = {
			position: "relative",
			display: "flex",
			flexDirection: "column",
			boxSizing: "border-box",
			width: "calc(100% - 8px)",
			height: "calc(100% - 8px)",
			margin: "8px 0px 0px 8px",
		}
		let headingStyle = {
			marginBottom: "8px",
		}
		let item = {
			IconUrl: iconUrl,
			Name: appArgument,
			Comment: comment
		}
		let itemStyle = {
			marginBottom: "8px",
		}
		let searchBoxStyle = {
			width: "calc(100% - 16px)",
			marginBottom: "3px",
		}
		let listStyle = {
			flex: "1",
		}
		return (
			<div style={contentStyle}>
				<div style={headingStyle}>Select an application to open:</div>
				<Item item={item} style={itemStyle}/>
				<SearchBox style={searchBoxStyle} onChange={this.onTermChange} searchTerm={this.state.searchTerm}/>
				<ItemList style={listStyle} items={apps} selected={selected} select={this.select} execute={this.execute}/>
			</div>
		)
	}
}

render(
	<AppChooser appArgument={gui.App.argv[0]} mimetypeId={gui.App.argv[1]}/>,
	document.getElementById('root')
);
