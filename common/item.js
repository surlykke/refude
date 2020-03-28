// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';

export let Item = props => {

    let {item, selected, onClick, onDoubleClick} = props;

    let style = {
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

    let iconStyle = {
        float: "left",
        marginRight: "6px"
    };

    Object.assign(iconStyle, item.iconStyle)

    let nameStyle = {
        overflow: "hidden",
        whiteSpace: "nowrap",
        marginRight: "6px",
    };

    let commentStyle = {
        fontSize: "0.8em",
    };

    return (
        <div id={item.url} style={style} onClick={() => onClick(item)} onDoubleClick={() => onDoubleClick(item)}>
            <img width="24px" height="24px" style={iconStyle} src={item.image} alt=""/>
            <div style={nameStyle}>{item.name}</div>
            <div style={commentStyle}>{item.comment}</div>
        </div>
    )
}
