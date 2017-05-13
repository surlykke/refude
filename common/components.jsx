import React from 'react';
import ReactDom from 'react-dom'

class List extends React.Component {
	constructor(props) {
		super(props)
	}

	componentDidUpdate() {
		// Scroll selected app into view
		if (this.props.selected) {
			let {top: listTop, bottom: listBottom} = this.refs["list"].getBoundingClientRect()
			let {top: selectedTop, bottom: selectedBottom} = this.refs[this.props.selected.url].getBoundingClientRect()
			if (selectedTop < listTop) this.refs["list"].scrollTop -=  (listTop - selectedTop + 15)
			else if (selectedBottom > listBottom) this.refs["list"].scrollTop += (selectedBottom - listBottom + 15)
		}
	}

	selected = (item, selected) => item === this.props.selected ? " selected" : ""

	render = () => {
		let {listOfLists, select, extraClasses} = this.props
		if (!extraClasses) {
			extraClasses = item => ""
		}

		return (
		    <div className="list" ref="list">
				{listOfLists.map(pair => (
				    <div key={pair.desc} className="sublist">
						<div className="sublistheading">{pair.desc}</div>
						{pair.items.map(item => (
							<div key={item.url}
								 ref={item.url}
								 onClick={() => {select(item)}}
								 onDoubleClick={() => {select(item, true)}}
								 className={"line" + this.selected(item) + extraClasses(item)}>

							    <div className="line-icon">
							        <img src={item.IconUrl} height="22" width="22" alt=" "/>
							    </div>
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
