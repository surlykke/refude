import React from 'react'
import {render} from 'react-dom'
import {MakeCollection} from '../../common/resources'
import {doHttp} from '../../common/utils'

let Notification = (props) => {

	let dismiss = (event) => {
		console.log(event)
		doHttp(props.item.url, "DELETE")
		event.stopPropagation()
	}

	let notificationClicked = (event) => {
		console.log("notification clicked")
		doHttp(props.item.url + "?action=default", "POST")
		event.stopPropagation()
	}

	let {item} = props

	let crossStyle = {
		position: "absolute",
		top: "3px",
		right: "2px",
		width: "15px",
		height: "15px",
	}

	// dangerouslySetInnterHtml should be safe here - we rely on
	// RefudeNotificationsService to sanitize notification body
	return (
		<div className="notification" style={{position: "relative"}} onClick={notificationClicked}>
			<div className="notificationHeading">{item.Subject}</div>
			<div className="notificationBody" dangerouslySetInnerHTML={{__html: item.Body}} />

			{Object.keys(item.Actions).filter(k => k !== "default").map(k => {
				let buttonClicked = (event) => {
					console.log("button", k, "clicked")
					doHttp(props.item.url + "?action=" + k, "POST")
					event.stopPropagation()
				}

				return 	<input type="submit" value={item.Actions[k]} onClick={buttonClicked}/>
			})}

			<div style={crossStyle} onClick={dismiss}>
				<svg height="15px" width="15px" viewBox="0 0 100 100"
					 strokeLinecap="round" stroke="grey" strokeWidth="10" >
					<circle cx="50" cy="50" r="40" fill="none"/>
					<line x1="32" y1="32" x2="68" y2="68"/>
					<line x1="32" y1="68" x2="68" y2="32"/>
				</svg>
			</div>

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
