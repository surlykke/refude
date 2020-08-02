//Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.

import React from 'react';
import './Do.css'

export let DoItem = props => {

    let {res, selected, onClick, onDoubleClick} = props;
    let iconUrl = res.IconName ? `http://localhost:7938/icon?name=${res.IconName}&theme=oxygen` : '';

    let itemClassName = "item" + (selected ? " selected" : "")
    let iconClassName = "icon"
    if (res.Type === "window") {
        iconClassName += " window" 
        if (res.Data.States && res.Data.States.includes("_NET_WM_STATE_HIDDEN")) {
            iconClassName += " minimized"    
        }
    }

    console.log("itemClassName:", itemClassName)
    console.log("iconClassName:", iconClassName)
    return [ 
        <div key="body" id={res.Self} className={itemClassName}  onClick={onClick} onDoubleClick={onDoubleClick}> 
            {res.OnDelete &&
            <svg width="10" height="10" viewBox="0 0 100 100" style={{position: "absolute", top: "2px", right: "2px"}}>
                <g fillOpacity="0" strokeWidth="12" stroke="black">
                    <line x1="25" y1="25" x2="75" y2="75"/>
                    <line x1="25" y1="75" x2="75" y2="25"/>
                </g>
                <circle cx="50" cy="50" r="48" stroke="black" strokeWidth="4" fill="none" />
            </svg>}

            {res.Actions &&
            <svg width="18" height="6" viewBox="0 0 300 100" style={{position: "absolute", zAxis: "10", bottom: "2px", right: "5px"}}>
                <circle cx="50" cy="50" r="30" stroke="gray" strokeWidth="4" fill="gray" />
                <circle cx="150" cy="50" r="30" stroke="gray" strokeWidth="4" fill="gray" />
                <circle cx="250" cy="50" r="30" stroke="gray" strokeWidth="4" fill="gray" />
            </svg>}
            <img className={iconClassName} width="24px" height="24px"  src={iconUrl} alt="" />
            <div className="name">{res.Title}</div>
            <div className="comment">{res.Comment}</div>
        </div>
    ]
}

