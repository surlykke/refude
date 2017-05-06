import React from 'react';
import ReactDom from 'react-dom'

let Mimetype = props =>
	<div className="searchBox">
		<div className="heading2">Select an application to open:</div>
	    <div className="line-icon">
	        <img src={props.mimetype ? props.mimetype.IconUrl : ""} height="32" width="32" alt=" "/>
		</div>
	    <div className="line-title">{props.mimetype ? props.mimetype.Comment : "?"}</div>
	    <div className="line-comment">{props.mimetype ? props.mimetype.Type + "/" + props.mimetype.Subtype : "?"}</div>
	</div>

let Applist = props =>
    <div className="list">
		{props.applist.map(pair => (<Appsublist key={pair.desc} desc={pair.desc} apps={pair.apps}/>))}
	</div>

let Appsublist = props =>
	<div >
		<div>{props.desc}</div>
		{props.apps.map(app => (<App key={app.url} app={app}/>))}
	</div>

let appClasses = (res, selected) => "line" + (selected ? " selected" : "")

let App = props =>
	<div onClick={() => {props.select(props.app)}}
		 onDoubleClick={() => {props.select(props.app, true)}}
		 className="line">

	    <div className="line-icon">
	        <img src={props.app.IconUrl} height="32" width="32" alt=" "/>
	    </div>
	    <div className="line-title">{props.app.Name}</div>
	    <div className="line-comment">{props.app.Comment}</div>
	</div>

export {Mimetype, Applist}
