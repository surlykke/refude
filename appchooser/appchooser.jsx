import React from 'react';
import {render} from 'react-dom';
import {doHttp, iconServiceUrl} from '../common/utils'
import {MakeCollection} from '../common/resources'
import {Argument} from "./components"
import {List} from "../common/components"


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
		this.apps = MakeCollection("desktop-service", "/applications	", this.update)
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
	update = () => {
		if (! this.updatePending) {
			this.updatePending = true
			setTimeout(this.updateHelper, 20)
		}
	}

	updateHelper = () => {
		let listOfLists = this.mimetypeIds.concat(["other"]).map(id => ({id:id, desc: "", items: []}))
		let appsTakingArgs = this.apps.filter(app => app.Actions["_default"]["Exec"].match(/%f|%F|%u|%U/))
		appsTakingArgs.forEach(app => {
		 	listOfLists.find(t => app.Mimetypes.includes(t.id) || t.id === 'other').items.push(app)
		})
		listOfLists = listOfLists.filter(t => t.items.length > 0)
		if (listOfLists.length > 1) {
			listOfLists.forEach(l => {
				l.desc = l.id === 'other' ? "Other applications" :
				   			   			    "Applications that handle " + this.mimetypes[l.id].Comment
			})
		}
		this.allItems = []
		listOfLists.forEach(t => {this.allItems.push(...t.items)})
		this.setState({listOfLists: listOfLists, selected: this.allItems[0]})
		this.updatePending = false
	}

	select = (item, execute) => {
		this.setState({selected: item})
		if (execute) this.execute()
	}

	execute = () => {
		let app = this.state.selected
		console.log("this.state.remember: ", this.state.remember)
		if (app) {
			if (this.state.remember) {
				let mimetypeUrl = "http://localhost:7938/desktop-service/mimetypes/" + mimetypeId
				doHttp(mimetypeUrl, "POST", {DefaultApplication: app.Id})
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

	checkboxChanged = event => {
		this.setState({remember: event.target.checked})
	}

	render = () => {
		let {appArgument, iconUrl, comment, listOfLists, selected} = this.state
		return (
			<div className=" content">
				<div className="topdown">
					<div className="heading2">Select an application to open:</div>
					<Argument appArgument={appArgument} iconUrl={iconUrl} comment={comment}/>
					<div> <input type="checkbox" onChange={this.checkboxChanged}/>Remember my decision</div>
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
