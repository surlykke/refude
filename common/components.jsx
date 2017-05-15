import React from 'react';
import ReactDom from 'react-dom'

class List extends React.Component {
	constructor(props) {
		super(props)
	}

	componentDidUpdate() {
		// Scroll selected app into view
		console.log("Scrolling into view...")
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
		console.log("List render, state: ", this.state)
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
