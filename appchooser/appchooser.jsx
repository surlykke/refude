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
		this.state = {
			appArgument: appArgument,
			iconUrl: iconServiceUrl(["unknown"], 32),
			comment: mimetypeId,
			listOfLists: []}
		this.allItems = []
	}

	componentDidMount() {
		this.fetch(mimetypeId)
		appsProxy.subscribe(url => {this.update()})
		document.body.addEventListener("keydown", this.onKeyDown)
	}

	fetch = (id) => {
		let url = "http://localhost:7938/desktop-service/mimetype/" + id
		doHttp(url).then(mimetype => {
			if (! this.mimetypeIds.includes(id)) {
				console.log("pushing", id, ", mimetype: ", mimetype)
				this.mimetypeIds.push(id)
				mimetype.url = url
				mimetype.IconUrl = iconServiceUrl([mimetype.IconName, mimetype.GenericIcon])
				this.mimetypes[id] = mimetype
				mimetype.SubClassOf.forEach(subId => { this.fetch(subId)})
				if (id === mimetypeId) {
					console.log("setState: ", {iconUrl: mimetype.IconUrl, comment: mimetype.Comment})
					this.setState({iconUrl: mimetype.IconUrl, comment: mimetype.Comment})
				}
				this.update()
			}
		})
	}



	// We get a lot of events from appsProxy, so we collect to, at most, one update pr 20 ms
	update = () => {
		if (! this.updatePending) {
			this.updatePending = true
			setTimeout(() => {
				console.log("update, mimetypeIds: ", this.mimetypeIds)
				let listOfLists = this.mimetypeIds.concat(["other"]).map(id => ({id:id, desc: "", items: []}))
				appsProxy.resources().filter(app => app.Actions["_default"]["Exec"].match(/%f|%F|%u|%U/))
				                     .forEach(app => {
				 	listOfLists.find(t => app.Mimetypes.includes(t.id) || t.id === 'other').items.push(app)
				})
				listOfLists = listOfLists.filter(t => t.items.length > 0)
				if (listOfLists.length > 1) {
					listOfLists.forEach(l => {l.desc = l.id === 'other' ?
					                                   "Other applications" :
							   			   			   "Applications that handle " + this.mimetypes[l.id].Comment})
				}
				this.allItems = []
				listOfLists.forEach(t => {this.allItems.push(...t.items)})
				this.setState({listOfLists: listOfLists, selected: this.allItems[0]})
				this.updatePending = false
			}, 20)
		}
	}

	select = (item, execute) => {
		this.setState({selected: item})
		if (execute) this.execute()
	}

	execute = () => {
		let app = this.state.selected
		let remember = this.refs.remember.value === "on"
		let mimetype = this.state.mimetype
		if (app) {
			if (remember && mimetype) {
				console.log("PATCHING to ", this.state.mimetype.url)
				let defaultApps = [app.Id, ...mimetype.DefaultApplications.filter(id => app.Id !== id)]
				doHttp(mimetype.url, "PATCH", {DefaultApplications: defaultApps})
			}

			doHttp(app.url, "POST", {Arguments: [appArgument]}).then(response => {gui.App.quit()})
		}
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
	        " " : () => {this.select(this.state.selected, true)}
		}[key]

		if (op) {
			op()
			event.preventDefault()
		}
	}

	extraClasses = item => "app"


	render = () => {
		let {appArgument, iconUrl, comment, listOfLists, selected} = this.state
		return (
			<div className=" content">
				<div className="topdown">
					<div className="heading2">Select an application to open:</div>
					<Argument appArgument={appArgument} iconUrl={iconUrl} comment={comment}/>
					<div> <input type="checkbox" ref="remember"/>Remember my decision</div>
					<div className="hr"></div>
					<List listOfLists={listOfLists} select={this.select} selected={selected} extraClasses={this.extraClasses}/>
				</div>
			</div>
		)
	}
}

render(
	<AppChooser appArgument={gui.App.argv[0]} mimetypeId={gui.App.argv[1]}/>,
	document.getElementById('root')
);
