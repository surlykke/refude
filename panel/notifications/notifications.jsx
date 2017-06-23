import React from 'react'
import {render} from 'react-dom'
import {MakeCollection} from '../../common/resource-collection'
import {doHttp} from '../../common/utils'

let Notification = (props) => {

	let dismiss = (event) => {
		doHttp(props.item.url, "DELETE")
	}

	return (
		<div className="notification" onClick={dismiss}>
			<div className="notificationHeading">{props.item.Subject}</div>
			<div className="notificationBody">{props.item.Body}</div>
		</div>)
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
