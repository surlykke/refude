import React from 'react';
import ReactDom from 'react-dom'

let SearchBox = props =>
    <div className="searchInput" onChange={props.onTermChange} >
        <input type="search" autoFocus value={props.searchTerm}/>
    </div>


let CommandList = props =>
    <div className="list" id="contentBox">
		{ props.windows.map(res => (
			<Command key={res.url} res={res} selected={props.selected} select={props.select}/>)
		)}
		{props.apps.map(res => (
			<Command key={res.url} res={res} selected={props.selected} select={props.select}/>)
		)}
	</div>

let isWindow = res => res.X !== undefined
let isMinimized = res =>  (res.States || []).includes("_NET_WM_STATE_HIDDEN")


let commandClasses = (res, selected) =>
	"line" + (res === selected ? " selected" : "") + (isWindow(res) ? " shadow" : "") + (isMinimized(res) ? " dimmed" : "")

let iconSize = res => isMinimized(res) ? 20 : 32


let Command = props =>
	<div onClick={() => {props.select(props.res)}}
		 onDoubleClick={() => {props.select(props.res, true)}}
		 className={commandClasses(props.res, props.selected)}>

	    <div className="line-icon" style={{paddingLeft: 32 - iconSize(props.res)}}>
	        <img src={props.res.IconUrl} height={iconSize(props.res)} width={iconSize(props.res)} alt=" "/>
	    </div>
	    <div className="line-title">{props.res.Name}</div>
	    <div className="line-comment">{props.res.Comment}</div>
	</div>


let zIndex = (win, index, selected) => win === selected ? 1 : -index
let fillOpacity = (win, selected) => win === selected ? "0.1" : "0.05"

let Windows = props =>
	<div id="disp" className="display">
		<svg viewBox="0 0 1920 1080" >
		{props.windows.map((win, index) => (
			<g key={win.url} z={zIndex(win, index, props.selected)} font-family="Verdana" fillOpacity={fillOpacity(win, props.selected)}>
			    <rect x={win.X} y={win.Y} width={win.W} height={win.H} stroke="black" />
				<rect x={win.X} y={win.Y} width={win.W} height="40" fill="lightblue" fillOpacity="1"/>
				<img xlinkHref={win.IconUrl} x={win.X} y={win.Y} width="40" height="40"/>
				<text x={win.X + 80} y={win.Y + 30} fontSize="27" stroke="black" alignmentBaseline="center">{win.Name}</text>
				<text x={win.X + win.W/2} y={win.Y + win.H/2} textAnchor="middle" alignmentBaseline="center" fontSize="120" stroke="black" fill="#000000">{index + 1}</text>
			</g>
		))}
		</svg>
	</div>

export {SearchBox, CommandList, Windows}
