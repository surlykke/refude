import React from 'react';
import ReactDom from 'react-dom'

let Item = props => {

	let {item, selected, select, execute} = props

	let style = {
		marginRight: "5px",
		padding: "4px",
	    verticalAlign: "top",
	    overflow: "hidden",
		height: "30px",
	}

	Object.assign(style, props.style)

	if (selected) {
		Object.assign(style, {
			border: "solid black 2px",
		    borderRadius: "5px",
	    	boxShadow: "1px 1px 1px #888888",
		})
	}

	let iconStyle =  {
		float: "left",
    	marginRight: "6px",
	}

	if (item.States) {  // Its a window
		Object.assign(iconStyle, {
	    	WebkitFilter: "drop-shadow(5px 5px 3px grey)",
		    overflow: "visible"
		})

		if (item.States.includes("_NET_WM_STATE_HIDDEN")) {
			Object.assign(item, {
				marginLeft: "14px",
				width: "18px",
				height: "18px",
			    opacity: "0.4"
			})
		}
	}

	Object.assign(iconStyle, props.item.extraIconStyle)

	let nameStyle = {
	    overflow: "hidden",
	    whiteSpace: "nowrap",
	    marginRight: "6px",
	}

	let commentStyle = {
    	fontSize: "0.8em",
	}

	Object.assign(commentStyle, nameStyle)

	return (
		<div id={props.item.url} style={style} onClick={() => {select(item)}} onDoubleClick={() => {execute(item)}}>
			<img width="24px" height="24px" style={iconStyle} src={item.IconUrl}/>
		    <div style={nameStyle}>{item.Name}</div>
		    <div style={commentStyle}>{item.Comment}</div>
		</div>
	)
}

export {Item}
