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
		this.apps = MakeCollection("desktop-service", "/applications	", this.scheduleUpdate)
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
		let term = this.state.searchTerm.toUpperCase().trim()
		let reg = isUrl ? /%u/i : /%f|%u/i
		let apps = this.apps.
			filter(app => app.Actions["_default"]["Exec"].match(reg)).
			filter(app => app.Name.toUpperCase().includes(term))

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

		if (!this.userHasMadeSelection || !this.state.selected || !apps.includes(this.state.selected)) {
			this.setState({selected: apps[0]})
		}

		this.setState({apps: apps})
		this.updatePending = false
	}

	select = (app) => {
		this.userHasMadeSelection = true
		this.setState({selected: app})
	}

	run = (app) => {
		this.select(app)
		if (this.state.useAsDefault)  {
			let mimetypeUrl = "http://localhost:7938/desktop-service/mimetypes/" + mimetypeId
			console.log("set default ", app.id, ", on ", mimetypeId)
			doHttp(mimetypeUrl, "POST", {DefaultApplication: app.Id}).then(resp => {
				console.log("Running ", app.url, appArgument)
				doHttp(app.url, "POST", {Arguments: [appArgument]}).then(resp => {
					gui.App.quit()
				})
			})
		} else {
			doHttp(app.url, "POST", {Arguments: [appArgument]}).then(resp => {
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
		else if (key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) this.run(this.state.selected)
		else if (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey) this.run(this.state.selected)
		else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) gui.App.quit()
		else return;
		event.stopPropagation()
	}

	move = (down) => {
		let {apps, selected} = this.state
		let index = apps.indexOf(selected)
		if (index > -1) {
			index = (index + apps.length + (down ? 1 : -1)) % apps.length
		}
		this.select(apps[index])
	}

	onTermChange = (event) => {
		this.setState({searchTerm: event.target.value})
		this.scheduleUpdate()
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

		return (
			<div style={styles.content} onKeyDown={this.onKeyDown}>
				<div style={styles.heading}>Select an application to open:</div>
				<Item item={item} style={styles.item}/>
				<SearchBox style={styles.searchBox} onChange={this.onTermChange} searchTerm={this.state.searchTerm}/>
				<ItemList style={styles.list} items={apps} selected={selected} select={this.select} execute={this.run}/>
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
