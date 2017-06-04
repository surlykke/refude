import React from 'react'
import {render} from 'react-dom'
import {MakeServiceProxy} from '../../common/service-proxy'

const statusNotifierItems = MakeServiceProxy("http://localhost:7938/statusnotifier-service/items/",
                                             "http://localhost:7938/statusnotifier-service/notify")

let NotifierItem = (props) => {
	let style = {
		display: "inline-block",
		background: "url(" + props.item.IconUrl + ")",
		backgroundSize: "contain",
		padding: "0px",
		margin: "0px",
		marginRight: "4px",
		fontSize: "1.3em",
		height: "100%",
		width: "20px"
	}
	return (<div className="notiferItem" style={style}/>)
}


class NotifierItems extends React.Component {
	constructor(props) {
		super(props)
		this.state = {items : []}
		console.log("constructor: this.state.items:", this.state.items)
	}

	componentDidMount() {
		console.log("Mount: this.state.items:", this.state.items)
		statusNotifierItems.subscribe(this.updateItems)
	}

	updateItems = () => {
		let items = statusNotifierItems.resources()
		this.setState({items: items})
		console.log("Update: this.state.items:", this.state.items)
	}

	render = () =>
		<div className="panel-plugin">
			{this.state.items.map((item) => (<NotifierItem key={item.id} item={item} /> ))}
		</div>
}

export {NotifierItems}
