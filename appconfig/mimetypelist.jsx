import React from 'react';
import ReactDom from 'react-dom'

let SearchBox = props =>
    <div className="searchInput" onChange={props.onTermChange} >
        <input type="search" autoFocus value={props.searchTerm}/>
    </div>


let MimetypeList = props =>
    <div className="list" id="contentBox">
		{ props.mimetypes.map(mimetype => (
			<Mimetype key={mimetype.url} mimetype={mimetype} selected={props.selected} select={props.select}/>)
		)}
	</div>

let mimetypeClasses = (mimetype, selected) => {
	return "line" +
	       (mimetype === selected ? " selected" : "") +
		   (mimetype.X !== undefined ? " shadow" : "") +
		   (mimetype.States && mimetype.States.includes("_NET_WM_STATE_HIDDEN") ? " dimmed" : "")
}

let Mimetype = props =>
	<div onClick={() => {props.select(props.mimetype)}}
		 onDoubleClick={() => {props.select(props.mimetype, true)}}
		 className={mimetypeClasses(props.mimetype, props.selected)}>

	    <div className="line-icon">
	        <img src={props.mimetype.IconUrl} height="32" width="32" alt=" "/>
	    </div>
	    <div className="line-title">{props.mimetype.Comment}</div>
	    <div className="line-comment">{props.mimetype.Type}/{props.mimetype.Subtype}</div>
	</div>


export {SearchBox, MimetypeList}
