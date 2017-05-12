import React from 'react';
import {render} from 'react-dom';
import {doHttp, iconServiceUrl} from '../common/utils'
import {MakeServiceProxy} from '../common/service-proxy'
import {Argument, Applist} from "./components"

let appsProxy = MakeServiceProxy("http://localhost:7938/desktop-service/applications",
                                 "http://localhost:7938/desktop-service/notify")

let gui = window.require('nw.gui')
if (gui.App.argv.length < 2) {
	gui.App.quit()
}
let appArgument = gui.App.argv[0]
let mimetypeId = gui.App.argv[1]

class AppChooser extends React.Component {
	constructor(props) {
		super(props)
		this.mimetypeIds = []
		this.mimetypes = new Map()
		this.state = {applist: []}
		this.allApps = []
	}

	componentDidMount() {
		console.log("fetch: ")
		this.fetch(mimetypeId)
		appsProxy.subscribe(url => {this.update()})
		document.body.addEventListener("keydown", this.onKeyDown)
	}

	fetch = (id) => {
		let url = "http://localhost:7938/desktop-service/mimetype/" + id
		doHttp(url).then(mimetype => {
			if (! this.mimetypeIds.includes(id)) {
				this.mimetypeIds.push(id)
				mimetype.IconUrl = iconServiceUrl([mimetype.IconName, mimetype.GenericIcon])
				this.mimetypes[id] = mimetype
				mimetype.SubClassOf.forEach(subId => { this.fetch(subId)})
				this.update()
			}
		})
	}

	// We get a lot of events from appsProxy, so we collect to, at most, one update pr 20 ms
	update = () => {
		if (! this.updatePending) {
			this.updatePending = true
			setTimeout(() => {
				let applist = this.mimetypeIds.concat(["other"]).map(id => {
					return {
						id: id,
						desc: id !== "other" ? "Applications that handle " + this.mimetypes[id].Comment : "Other applications",
						apps: []
					}
				})

				let takesArgs = app => app.Actions["_default"]["Exec"].match(/%f|%F|%u|%U/)
				appsProxy.resources().filter(takesArgs).forEach(app => {
				 	applist.find(t => app.Mimetypes.includes(t.id) || t.id === 'other').apps.push(app)
				})
				applist = applist.filter(t => t.apps.length > 0)

				this.allApps = []
				applist.forEach(t => {this.allApps.push(...t.apps)})
				this.setState({appArgument: appArgument, mimetype: this.mimetypes[mimetypeId], selected: this.allApps[0], applist: applist})
				this.updatePending = false
			}, 20)
		}
	}

	select = (app, done) => {
		this.setState({selected: app})
		if (done) {
			this.execute()
		}
	}

	move = up => {
		let i = this.allApps.indexOf(this.state.selected)
		i = (i + (up ? -1 : 1) + this.allApps.length) % this.allApps.length
		this.setState({selected: this.allApps[i]})
	}

	execute = () => {
		if (this.state.selected) {
			let url = this.state.selected.url + "?arg=" + encodeURIComponent(appArgument)
			doHttp(url, "POST").then(response => {gui.App.quit()})
		}
	}

	onKeyDown = (event) => {
		let {key, ctrlKey, shiftKey, altKey, metaKey} = event
		let op = {
			Tab: () => {this.move(shiftKey)},
	        ArrowDown : () => {this.move()},
	        ArrowUp :  () => {this.move(true)},
	        Enter : () => {this.execute()},
	        " " : () => {this.execute()}
		}[key]

		if (op) {
			op()
			event.preventDefault()
		}
	}


	appClasses = (app, selected) => "line" + (app === selected ? " selected" : "")

	render = () => {
		let {mimetype, applist, selected} = this.state
		return (
			<div className=" content" onKeyDown={this.onKeyDown}>
				<div className="topdown">
					<div className="heading2">Select an application to open:</div>
					<Argument appArgument={appArgument} mimetypeId={mimetypeId} mimetype={mimetype}/>
					<div> <input type="checkbox"/>Remember</div>
					<div className="hr"></div>
					<Applist applist={applist} selected={selected} select={this.select}/>
				</div>
			</div>
		)
	}
}

render(
	<AppChooser appArgument={gui.App.argv[0]} mimetypeId={gui.App.argv[1]}/>,
	document.getElementById('root')
);
