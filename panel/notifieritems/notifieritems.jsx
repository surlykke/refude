import React from 'react'
import {render} from 'react-dom'
import {MakeCollection} from '../../common/resources'
import {doHttp} from '../../common/utils'

let NotifierItem = (props) => {
	let style = {
		display: "inline-block",
		backgroundImage: 'url(' + props.item.IconUrl + ')',
		backgroundPositionX: 'center',
		backgroundPositionY: 'center',
		backgroundRepeat: 'no-repeat',
		backgroundSize: "contain",
		padding: "0px",
		margin: "0px",
		marginRight: "4px",
		fontSize: "1.3em",
		height: "100%",
		width: "20px"
	}

	let getXY = (event) => {
		return  {
			x: Math.round(event.view.devicePixelRatio * event.screenX),
			y: Math.round(event.view.devicePixelRatio * event.screenY)
		}
	}

	let onClick = (event) => {
		event.persist()
		console.log(event)
		if (event.button === 0) {
			call("Activate", getXY(event))
		} else if (event.button === 1){
			call("SecondaryActivate", getXY(event))
		}
	}

	let onRightClick = (event) => {
		event.persist()
		call("ContextMenu", getXY(event))
		event.preventDefault()
	}

	let call = (method, xy) => {
		let url = props.item.url + `?method=${method}&x=${xy.x}&y=${xy.y}`
		console.log("Posting: ", url)
		doHttp(url, "POST")
	}

	return (<div className="notiferItem" style={style}  onClick={onClick} onContextMenu={onRightClick}/>)
}


class NotifierItems extends React.Component {
	constructor(props) {
		super(props)
		this.state = {items : []}
		this.onUpdated = props.onUpdated
		console.log("constructor: this.state.items:", this.state.items)
	}

	componentDidMount() {
		this.items = MakeCollection("statusnotifier-service", "/items", this.update)
	}

	componentDidUpdate() {
		this.onUpdated()
	}

	update = () => {
		this.setState({items: this.items.slice()})
	}

	render = () =>
		<div className="panel-plugin notifier-items">
			{this.state.items.map((item) => (<NotifierItem key={item.id} item={item} /> ))}
		</div>
}

export {NotifierItems}