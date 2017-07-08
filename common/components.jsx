// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';
import ReactDom from 'react-dom'

class List extends React.Component {
	constructor(props) {
		super(props)
	}

	componentDidUpdate() {
		// Scroll selected app into view
		if (this.selectedDiv) {
			let {top: listTop, bottom: listBottom} = this.listDiv.getBoundingClientRect()
			let {top: selectedTop, bottom: selectedBottom} = this.selectedDiv.getBoundingClientRect()
			if (selectedTop < listTop) this.listDiv.scrollTop -=  (listTop - selectedTop + 15)
			else if (selectedBottom > listBottom) this.listDiv.scrollTop += (selectedBottom - listBottom + 15)
		}
	}

	classes = (item, selected) => {
		let tmp = "line" + (item === selected ? " selected" : "")
		if (this.props.extraClasses) {
			tmp += " " + this.props.extraClasses(item)
		}
		return tmp
	}

	render = () => {
		let {listOfLists, select, selected} = this.props
		return (
		    <div className="list" ref={listDiv => this.listDiv = listDiv }>
				{listOfLists.map(pair => (
				    <div key={pair.desc} className="sublist">
						<div className="sublistheading">{pair.desc}</div>
						{pair.items.map(item => (
							<div key={item.url}
								 ref={div => {if (item == selected) this.selectedDiv = div}}
								 onClick={() => {select(item)}}
								 onDoubleClick={() => {select(item, true)}}
								 className={this.classes(item, selected)}>

							    <div className="line-icon"
									 style={{background: "url(" + item.IconUrl + ")	",
									         backgroundSize: "contain"}} />
							    <div className="line-title">{item.Name}</div>
							    <div className="line-comment">{item.Comment}</div>
							</div>
						))}
				    </div>
				))}
			</div>
		)
	}
}

export {List}
