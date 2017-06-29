import React from 'react'
import {render} from 'react-dom'
import {MakeCollection} from '../../common/resources'
import {doHttp} from '../../common/utils'

let Notification = (props) => {

	let dismiss = () => {
		doHttp(props.item.url, "DELETE")
	}

	let action = (id) => {
		doHttp(props.item.url + "?action=" + id)
	}

	let {item} = props
	console.log("Actions: ", item.Actions, ", keys: ", Object.keys(item.Actions))
	// dangerouslySetInnterHtml should be safe here - we rely on
	// RefudeNotificationsService to sanitize notification body
	return (
		<div className="notification" onClick={() => action("default")}>
			<div className="notificationHeading">{item.Subject}</div>
			<div
				className="notificationBody"
				dangerouslySetInnerHTML={{__html: item.Body}} />
			{Object.keys(item.Actions).filter(k => k !== "default").map(k =>
				<input type="submit" id={k} value={item.Actions[k]} onClick={() => {action(k)}}/>)
			}
		</div>
	 )
}


class Notifications extends React.Component {
	constructor(props) {
		super(props)
		this.state = {items : []}
		this.onUpdated = props.onUpdated
		console.log("constructor: this.state.items:", this.state.items)
	}

	componentDidMount() {
		this.notifications = MakeCollection("notifications-service", "/notifications", this.update)
	}

	componentDidUpdate() {
		this.onUpdated()
	}

	update = () => {
		console.log("update, collection: ", this.notifications)
		this.setState({items: this.notifications.slice()})
	}

	render = () =>
		<div className="notifications">
			{this.state.items.map(item =>
				(<Notification key={item.Id} item={item} /> ))}
		</div>
}

export {Notifications}
