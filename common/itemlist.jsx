// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
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
				if (selectedTop < listTop) listDiv.scrollTop -=  (listTop - selectedTop + 25)
				else if (selectedBottom > listBottom) listDiv.scrollTop += (selectedBottom - listBottom + 10)
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
			content.push(<Item key={item.url} item={item} selected={item === selected} select={select} execute={execute}/>)
		})
		return (
			<div id="itemListDiv" style={style}>
				{content}
			</div>
		)
	}
}

export {ItemList}
