import React from 'react';
import ReactDom from 'react-dom'

let Argument = props =>
	<div key="mimetype" className="mimetype">
		<div className="line-icon">
        	<img src={props.iconUrl} height="32" width="32" alt=" "/>
		</div>
	    <div className="line-title">{props.appArgument}</div>
	    <div className="line-comment">Type: {props.comment}</div>
	</div>


export {Argument}
