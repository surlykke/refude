import React from 'react';
import ReactDom from 'react-dom'
import {Item} from './item.jsx'

class ItemList extends React.Component {

	constructor(props) {
		super(props)
	}

	componentDidUpdate() {
		if (this.props.selected) {
			let selectedDiv = document.getElementById(this.props.selected.url)
			if (selectedDiv) {
				let listDiv = document.getElementById("itemListDiv")
				let {top: listTop, bottom: listBottom} = listDiv.getBoundingClientRect()
				let {top: selectedTop, bottom: selectedBottom} = selectedDiv.getBoundingClientRect()
				if (selectedTop < listTop) listDiv.scrollTop -=  (listTop - selectedTop + 15)
				else if (selectedBottom > listBottom) listDiv.scrollTop += (selectedBottom - listBottom + 15)
			}
		}
	}

	render = () => {
		let {items, selected, select, execute} = this.props
		let style = {
			overflow: "auto",
		}
		Object.assign(style, this.props.style)

		let headingStyle = {
			fontSize: "0.9em",
	    	color: "gray",
		    fontStyle: "italic",
			marginTop: "5px",
		    marginBottom: "3px",
		}

		let prevGroup
		let content = []
		items.forEach(item => {
			if (item.group !== prevGroup) {
				content.push(<div style={headingStyle}>{item.group}</div>)
				prevGroup = item.group
			}
			content.push(<Item item={item} selected={item === selected} select={select} execute={execute}/>)
		})
		return (
			<div id="itemListDiv" style={style}>
				{content}
			</div>
		)
	}
}

export {ItemList}
