// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
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
