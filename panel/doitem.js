//Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.

import React from 'react';

export let DoItem = props => {

    let {res, prevRes, selected, onClick, onDoubleClick} = props;

    let headingStyle = {
        fontSize: "0.9em",
        color: "gray",
        fontStyle: "italic",
        marginTop: "5px",
        marginBottom: "3px",
    };

    let style =  {
            marginRight: "5px",
            padding: "4px",
            verticalAlign: "top",
            overflow: "hidden",
            height: "30px",
    };

    if (selected) {
        Object.assign(style, {
            border: "solid black 2px",
            borderRadius: "5px",
            boxShadow: "1px 1px 1px #888888",
        })
    }

    let iconStyle =  {
        float: "left",
        marginRight: "6px"
    };

    if (res.Type === "window") {
        Object.assign(iconStyle, {
            WebkitFilter: "drop-shadow(5px 5px 3px grey)",
            overflow: "visible"
        });

        if (res.Data.States && res.Data.States.includes("_NET_WM_STATE_HIDDEN")) {
            Object.assign(iconStyle, {
                marginLeft: "10px",
                marginTop: "10px",
                width: "14px",
                height: "14px",
                opacity: "0.7"
            })
        }
    }

    let iconUrl = res.IconName ? `http://localhost:7938/icon?name=${res.IconName}&theme=oxygen` : '';

    let nameStyle = {
        overflow: "hidden",
        whiteSpace: "nowrap",
        marginRight: "6px",
    };

    let commentStyle = {
        fontSize: "0.8em",
    };
    
    return [ 
        !(prevRes && prevRes.Type === res.Type) &&
        <div style={headingStyle}>{res.Type}</div>,
        
        <div id={res.Self} style={style} onClick={onClick} onDoubleClick={onDoubleClick}>
            <img width="24px" height="24px" style={iconStyle} src={iconUrl} alt="" />
            <div style={nameStyle}>{res.Title}</div>
            <div style={commentStyle}>{res.Comment}</div>
        </div>
    ]
}

