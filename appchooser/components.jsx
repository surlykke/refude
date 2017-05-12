import React from 'react';
import ReactDom from 'react-dom'

let Argument = props =>
	<div key="mimetype" className="mimetype">
		<div className="line-icon">
        	<img src={props.mimetype ? props.mimetype.IconUrl : ""} height="32" width="32" alt=" "/>
		</div>
	    <div className="line-title">{props.appArgument}</div>
	    <div className="line-comment">Type: {props.mimetype ? props.mimetype.Comment: "?"}</div>
	</div>

class Applist extends React.Component {
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

	scrollIntoView() {

	}

	appClasses = (app, selected) => "line" + (app === selected ? " selected" : "")


	render = () => {
		let {applist, selected, select} = this.props
		return (
		    <div className="list" ref="list">
				{applist.map(pair => (
				    <div key={pair.desc} className="sublist">
						<div className="sublistheading">{pair.desc}</div>
						{pair.apps.map(app => (
							<div key={app.url}
								 ref={app.url}
								 onClick={() => {select(app)}}
								 onDoubleClick={() => {select(app, true)}}
								 className={this.appClasses(app, selected)}>

							    <div className="line-icon">
							        <img src={app.IconUrl} height="22" width="22" alt=" "/>
							    </div>
							    <div className="line-title">{app.Name}</div>
							    <div className="line-comment">{app.Comment}</div>
							</div>
						))}
				    </div>
				))}
			</div>
		)
	}
}



export {Argument, Applist}
