// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';
import ReactDom from 'react-dom'

let zIndex = (win, index, selectedSelf) => win.Self === selectedSelf ? 1 : -index
let fillOpacity = (win, selectedSelf) => win.Self === selectedSelf ? "0.1" : "0.05"

let Windows = props =>
	<div id="disp" style={props.style || {}}>
		<svg viewBox="0 0 3840 1080" >
		{props.windows.map((win, index) => (
			<g key={win.url} z={zIndex(win, index, props.selectedSelf)} fontFamily="Verdana" fillOpacity={fillOpacity(win, props.selectedSelf)}>
			    <rect x={win.X} y={win.Y} width={win.W} height={win.H} stroke="black" />
				<rect x={win.X} y={win.Y} width={win.W} height="40" fill="lightblue" fillOpacity="1"/>
				<img xlinkHref={win.IconUrl} x={win.X} y={win.Y} width="40" height="40"/>
				<text x={win.X + 80} y={win.Y + 30} fontSize="27" stroke="black" alignmentBaseline="center">{win.Name}</text>
				<text x={win.X + win.W/2} y={win.Y + win.H/2} textAnchor="middle" alignmentBaseline="center" fontSize="120" stroke="black" fill="#000000">{index + 1}</text>
			</g>
		))}
		</svg>
	</div>

export {Windows}
