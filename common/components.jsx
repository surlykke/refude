import React from 'react';
import ReactDom from 'react-dom'

class List extends React.Component {
	constructor(props) {
		super(props)
		this.getAllItems(props.listOfLists)
		this.state = {selected: this.allItems[0]}
	}

	componentWillReceiveProps(props) {
		this.getAllItems(props.listOfLists)
		if (! this.allItems.includes(this.state.selected)) {
			this.setState({selected: this.allItems[0]})
		}
	}

	getAllItems = listOfLists => {
		this.allItems = []
		listOfLists.forEach(t => this.allItems.push(...t.items))
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

	classes = item => {
		let tmp = "line" + (item === this.state.selected ? " selected" : "")
		if (this.props.extraClasses) {
			tmp += " " + this.props.extraClasses(item)
		}
		return tmp
	}

	select = (item, exec) => {
		this.setState({selected: item})
		if (exec) {
			this.props.execute(item)
		}
	}

	execute = () => {
		if (this.state.selected) {
			this.props.execute(this.state.selected)
		}
	}


	onKeyDown = (event) => {
		let {key, ctrlKey, shiftKey, altKey, metaKey} = event
		let op = {
			Tab: () => {this.move(shiftKey)},
	        ArrowDown : () => {this.move()},
	        ArrowUp :  () => {this.move(true)},
	        Enter : () => {this.execute()},
	        " " : () => {this.execute()}
		}[key]

		if (op) {
			op()
			event.preventDefault()
		}
	}

	move = up => {
		let i = this.allItems.indexOf(this.state.selected)
		i = (i + (up ? -1 : 1) + this.allItems.length) % this.allItems.length
		this.setState({selected: this.allItems[i]})
	}


	render = () => {
		let {listOfLists} = this.props

		return (
		    <div className="list" ref={listDiv => this.listDiv = listDiv }>
				{listOfLists.map(pair => (
				    <div key={pair.desc} className="sublist">
						<div className="sublistheading">{pair.desc}</div>
						{pair.items.map(item => (
							<div key={item.url}
								 ref={div => {if (item.url === this.state.selected.url) this.selectedDiv = div}}
								 onClick={() => {this.select(item)}}
								 onDoubleClick={() => {this.select(item, true)}}
								 className={this.classes(item)}>

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
