import React from 'react';
import ReactDom from 'react-dom'

let SearchBox = props => {
	let style = {
		boxSizing: "border-box",
		paddingRight: "5px",
	}
	Object.assign(style, props.style)

	let inputStyle =  {
	    width: "100%",
		height: "36px",
	    borderRadius: "5px",
	    outlineStyle: "none",
	}

	return (
		<div style={style}>
			<input style={inputStyle} type="search" onChange={props.onChange} value={props.searchTerm} autoFocus/>
		</div>
	)
}

export {SearchBox}
