// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
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
let isUrl = mimetypeId.startsWith("x-scheme-handler")

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
		this.apps = MakeCollection("desktop-service", "/applications", this.scheduleUpdate, (app, term) => {
			return app.Name.toUpperCase().includes(term) &&
				   (app.Exec.toUpperCase().includes("%F") || app.Exec.toUpperCase().includes("%U"))
		})
	}

	componentDidMount() {
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

	// We get a lot of events from this.apps, so we collect to, at most, one update pr 20 ms
	scheduleUpdate = () => {
		if (! this.updatePending) {
			this.updatePending = true
			setTimeout(this.update, 20)
		}
	}

	update = () => {
		let apps = []
		apps.push(...this.apps.filtered)
		apps.forEach(app => {
			let mimetypeId = this.mimetypeIds.find(id => app.Mimetypes.includes(id))
			if (mimetypeId) {
				app.group = "Applications that handle: '" + this.mimetypes[mimetypeId].Comment + "'"
				app.order = this.mimetypeIds.indexOf(mimetypeId)
			} else {
				app.group = "Other applications"
				app.order = this.mimetypeIds.length
			}
		})
		apps.sort((app1, app2) => app1.order !== app2.order ? app1.order - app2.order : app1.Name.localeCompare(app2.Name))

		if (!this.userHasMadeSelection ||
				!this.state.selected ||
				!apps.find(app => app === this.state.selected)) {
			this.setState({selected: apps[0]})
		}

		this.setState({apps: apps})
		this.updatePending = false
	}

	select = (url) => {
		this.userHasMadeSelection = true
		this.setState({selected: this.state.apps.find(app => app.url === url)})
	}

	run = (url) => {
		this.select(url)
		let app = this.state.apps.find(app => app.url === url)
		if (!app) return

		let launchUrl = url + "?arg=" + appArgument
		if (this.state.useAsDefault)  {
			let postUrl = "http://localhost:7938/desktop-service/mimetypes/" + mimetypeId + "?defaultApp=" + app.Id
			doHttp(postUrl, "POST").then(resp => {
				doHttp(launchUrl, "POST").then(resp => {
					gui.App.quit()
				})
			})
		} else {
			doHttp(launchUrl, "POST").then(resp => {
				gui.App.quit()
			})
		}
	}

	onKeyDown = (event) => {
		let {key, ctrlKey, shiftKey, altKey, metaKey} = event
		if      (key === "Tab" && !ctrlKey &&  shiftKey && !altKey && !metaKey) this.move(false)
		else if (key === "Tab" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true)
		else if (key === "ArrowUp" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(false)
		else if (key === "ArrowDown" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.move(true)
		else if (key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.run(this.state.selected.url)
		else if (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey) this.run(this.state.selected.url)
		else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) gui.App.quit()
		else return;
		event.stopPropagation()
	}

	move = (down) => {
		let {apps, selected} = this.state
		let index = apps.indexOf(selected)
		if (index > -1) {
			index = (index + apps.length + (down ? 1 : -1)) % apps.length
			this.select(apps[index].url)
		}
	}

	onTermChange = (event) => {
		this.apps.setterm(event.target.value)
		this.setState({searchTerm: event.target.value})
	}

	render = () => {
		let {iconUrl, comment, apps, selected, confirm} = this.state

		let styles = {
			content: {
				position: "relative",
				display: "flex",
				flexDirection: "column",
				boxSizing: "border-box",
				width: "100%",
				height: "100%",
				padding: "8px",
				//margin: "8px 8px 0px 8px",
			},
			heading: {
				marginBottom: "8px",
			},
			item: {
				marginBottom: "8px",
			},
			searchBox: {
				width: "calc(100% - 16px)",
				marginBottom: "3px",
			},
			list: {
				flex: "1",
				paddingBottom: "80px",
			},
			useAsDefault: {
				boxSizing: "border-box",
				width: "calc(100% - 16px)",
				paddingTop: "12px",
				paddingBottom: "8px",
				display: "flex",
				borderRadius: "5px",
				backgroundColor: "rgba(245,245,245,1)",
			},
		}

		let item = {
			IconUrl: iconUrl,
			Name: appArgument,
			Comment: comment
		}

		let selectedUrl = selected ? selected.url : undefined

		return (
			<div style={styles.content} onKeyDown={this.onKeyDown}>
				<div style={styles.heading}>Select an application to open:</div>
				<Item item={item} style={styles.item}/>
				<SearchBox style={styles.searchBox} onChange={this.onTermChange} searchTerm={this.state.searchTerm}/>
				<ItemList style={styles.list} items={apps} selectedUrl={selectedUrl} select={this.select} execute={this.run}/>
				<div style={{height: "8px"}}/>
				{	this.state.selected &&
					<div style={styles.useAsDefault}>
						<input id="checkbox"
							   type="checkbox"
							   value={this.state.useAsDefault}
							   onChange={(evt) => {this.setState({useAsDefault: evt.target.checked})}}/>
						<label htmlFor="checkbox" accessKey="M">
							Always use <em>{this.state.selected.Name}</em> to open {isUrl ? "urls" : "files"} of this type?
						</label>
					</div>
				}
			</div>
		)
	}
}

render(
	<AppChooser appArgument={gui.App.argv[0]} mimetypeId={gui.App.argv[1]}/>,
	document.getElementById('root')
);
