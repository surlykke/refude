import React from 'react';
import {render} from 'react-dom';
import {doHttp} from '../common/utils'
import {MakeServiceProxy} from '../common/service-proxy'
import {Mimetype, Applist} from "./components"

let appsProxy = MakeServiceProxy("http://localhost:7938/desktop-service/applications",
                                 "http://localhost:7938/desktop-service/notify")


class AppChooser extends React.Component {
	constructor(props) {
		super(props)
		let gui =
		this.mimetypeId =  window.require('nw.gui').App.argv[0]
		console.log("mimetypeId:", this.mimetypeId)
		let err = ""
		if (!this.mimetypeId) {
			err = "Usage: appchooser mimetype [uri]"
		}
		else {
			this.mimetypeIds = []
			this.mimetypes = new Map()
		}
		this.state = {applist: [], err: err}
	}


	componentDidMount() {
		console.log("Fetching...")
		this.fetch(this.mimetypeId)
		appsProxy.subscribe(url => {this.requestUpdate()})
	}

	fetch = (mimetypeId) => {
		let url = "http://localhost:7938/desktop-service/mimetype/" + mimetypeId
		console.log("Fetching ", url)
		doHttp(url).then(mimetype => {
			console.log("Got: ", mimetype)
			let mimetypeId = mimetype.Type + "/" + mimetype.Subtype
			if (! this.mimetypeIds.includes(mimetypeId)) {
				this.mimetypeIds.push(mimetypeId)
				this.mimetypes[mimetypeId] = mimetype
				mimetype.SubClassOf.forEach(id => { this.fetch(id)})
				this.requestUpdate()
			}
		})
	}

	requestUpdate = () => {
		if (! this.updatePending) {
			this.updatePending = true
			setTimeout(() => {this.update(); this.updatePending = false}, 20)
		}
	}

	update = () => {
		console.log("Updating")
		let apps = new Map()
		this.mimetypeIds.forEach(mimetypeId => {apps[mimetypeId] = []})
		apps["other"] = []

		appsProxy.resources().forEach(app => {
			let id = this.mimetypeIds.find(id => app.Mimetypes.includes(id)) || "other"
			apps[id].push(app)
		})

		let applist = []
		this.mimetypeIds.forEach(mimetypeId => {
			applist.push({desc: "Applications that handle " + this.mimetypes[mimetypeId].Comment, apps: apps[mimetypeId]})
		})
		applist.push({desc: "Other applications", apps: apps["other"]})

		this.setState({mimetype: this.mimetypes[this.mimetypeId], applist: applist})
		console.log("State now: ", this.state)
	}


	render = () => {
		return (
			<div className="content">
				<div>
					<Mimetype mimetype={this.state.mimetype}/>
					<Applist applist={this.state.applist}/>
				</div>
			</div>
		)
	}
}

render(
	<AppChooser/>,
	document.getElementById('root')
);
