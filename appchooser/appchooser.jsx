import React from 'react';
import {render} from 'react-dom';
import {doHttp, iconServiceUrl} from '../common/utils'
import {MakeServiceProxy} from '../common/service-proxy'
import {Argument} from "./components"
import {List} from "../common/components"

let appsProxy = MakeServiceProxy("http://localhost:7938/desktop-service/applications",
                                 "http://localhost:7938/desktop-service/notify")

let gui = window.require('nw.gui')
let appArgument = gui.App.argv[0]
let mimetypeId = gui.App.argv[1]

class AppChooser extends React.Component {
	constructor(props) {
		super(props)
		this.mimetypeIds = []
		this.mimetypes = new Map()
		this.state = {appArgument: appArgument, mimetypeId: mimetypeId, listOfLists: []}
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
				mimetype.url = url
				mimetype.IconUrl = iconServiceUrl([mimetype.IconName, mimetype.GenericIcon])
				this.mimetypes[id] = mimetype
				mimetype.SubClassOf.forEach(subId => { this.fetch(subId)})
				if (id === this.state.mimetypeId) {
					this.setState({mimetype: mimetype})
				}
				this.update()
			}
		})
	}


	desc = id => id === "other" ? "Other applications" : "Applications that handle " + this.mimetypes[id].Comment
	takesArgs = app => app.Actions["_default"]["Exec"].match(/%f|%F|%u|%U/)

	// We get a lot of events from appsProxy, so we collect to, at most, one update pr 20 ms
	update = () => {
		if (! this.updatePending) {
			this.updatePending = true
			setTimeout(() => {
				let listOfLists = this.mimetypeIds.concat(["other"]).map(id => {
					return { id: id, desc: this.desc(id), items: [] }
				})

				appsProxy.resources().filter(this.takesArgs).forEach(app => {
				 	listOfLists.find(t => app.Mimetypes.includes(t.id) || t.id === 'other').items.push(app)
				})
				listOfLists = listOfLists.filter(t => t.items.length > 0)

				this.setState({listOfLists: listOfLists})
				this.updatePending = false
			}, 20)
		}
	}

	execute = app => {
		if (this.refs.remember.value === "on" && this.state.mimetype) {
			console.log("PATCHING to ", this.state.mimetype.url)
			let defaultApps = [app.Id, ...this.state.mimetype.DefaultApplications.filter(id => app.Id !== id)]
			doHttp(this.state.mimetype.url, "PATCH", {DefaultApplications: defaultApps})
		}

		doHttp(app.url, "POST", {Arguments: [appArgument]}).then(response => {gui.App.quit()})
	}

	onKeyDown = (event) => {
		console.log("onKeyDown: ", event)
		if (["Tab", "ArrowDown", "ArrowUp", "Enter", " "].includes(event.key)) {
			this.list.onKeyDown(event)
		}
	}

	render = () => {
		let {mimetype, listOfLists, selected} = this.state
		return (
			<div className=" content" onKeyDown={this.onKeyDown}>
				<div className="topdown">
					<div className="heading2">Select an application to open:</div>
					<Argument appArgument={appArgument} mimetypeId={mimetypeId} mimetype={mimetype}/>
					<div> <input type="checkbox" ref="remember"/>Remember</div>
					<div className="hr"></div>
					<List listOfLists={listOfLists} execute={this.execute} ref={list => this.list = list}/>
				</div>
			</div>
		)
	}
}

render(
	<AppChooser appArgument={gui.App.argv[0]} mimetypeId={gui.App.argv[1]}/>,
	document.getElementById('root')
);
