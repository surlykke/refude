import React from 'react';
import ReactDom from 'react-dom'

let SearchBox = props =>
    <div className="searchInput" onChange={props.onTermChange} >
        <input type="search" autoFocus value={props.searchTerm}/>
    </div>

let classNames = (res, selected) => {
	console.log("classNames(", res, ", ", selected, ")")
	return "line" + (res === selected ? " selected" : "")
}

let CommandList = props =>
    <div className="list" id="contentBox">
		{ props.windows.map(res => (
			<Command key={res.url} res={res} selected={props.selected} select={props.select}/>)
		)}
		{props.apps.map(res => (
			<Command key={res.url} res={res} selected={props.selected} select={props.select}/>)
		)}
	</div>

let classes = (res, selected) => {
	return "line" + (res === selected ? " selected" : "")
}

let Command = props =>
	<div onClick={() => {props.select(props.res)}}
		 onDoubleClick={() => {props.select(props.res, true)}}
		 className={classes(props.res, props.selected)}>

	    <div className="line-icon">
	        <img src={props.res.IconUrl} height="32" width="32" alt=" "/>
	    </div>
	    <div className="line-title">{props.res.Name}</div>
	    <div className="line-comment">{props.res.Comment}</div>
	</div>


let zIndex = (win, index, selected) => win === selected ? 1 : -index
let fillOpacity = (win, selected) => win === selected ? "0.3" : "0.05"

let Windows = props =>
	<div id="disp" className="display">
		<svg viewBox="0 0 1920 1080" >
		{props.windows.map((win, index) => (
			<g key={win.url} z={zIndex(win, index, props.selected)} fillOpacity={fillOpacity(win, props.selected)}>
			    <rect x={win.X} y={win.Y} width={win.W} height={win.H} stroke="black"/>
				<text x={win.X + win.W/2} y={win.Y + win.H/2} textAnchor="middle" alignmentBaseline="center" fontSize="120">{index}</text>
			</g>
		))}
		</svg>
	</div>

export {SearchBox, CommandList, Windows}
