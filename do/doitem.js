//Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.

import React from 'react';

export let DoItem = props => {

    let {res, selected, onClick, onDoubleClick} = props;

    let style =  {
        position: "relative",
        marginRight: "5px",
        padding: "4px",
        paddingRight: "18px",
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

    if (res.OtherActions) {
        Object.assign(commentStyle, {width: "calc(100% - 24px)"})
    }

    return [ 
        <div key="body" id={res.Self} style={style} onClick={onClick} onDoubleClick={onDoubleClick}> 
            {res.OnDelete &&
            <svg width="10" height="10" viewBox="0 0 100 100" style={{position: "absolute", top: "2px", right: "2px"}}>
                <g fillOpacity="0" strokeWidth="12" stroke="black">
                    <line x1="25" y1="25" x2="75" y2="75"/>
                    <line x1="25" y1="75" x2="75" y2="25"/>
                </g>
                <circle cx="50" cy="50" r="48" stroke="black" strokeWidth="4" fill="none" />
            </svg>}

            {res.OtherActions &&
            <svg width="18" height="6" viewBox="0 0 300 100" style={{position: "absolute", zAxis: "10", bottom: "2px", right: "5px"}}>
                <circle cx="50" cy="50" r="30" stroke="gray" strokeWidth="4" fill="gray" />
                <circle cx="150" cy="50" r="30" stroke="gray" strokeWidth="4" fill="gray" />
                <circle cx="250" cy="50" r="30" stroke="gray" strokeWidth="4" fill="gray" />
            </svg>}
            <img width="24px" height="24px" style={iconStyle} src={iconUrl} alt="" />
            <div style={nameStyle}>{res.Title}</div>
            <div style={commentStyle}>{res.Comment}</div>
        </div>
    ]
}

