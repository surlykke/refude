// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';

let Item = props => {

    let {item, selected, select, execute} = props;

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
        marginRight: "6px",
    };

    let nameStyle = {
        overflow: "hidden",
        whiteSpace: "nowrap",
        marginRight: "6px",
    };

    let commentStyle = {
        fontSize: "0.8em",
    };

    return (
        <div id={props.item._self} style={style} onClick={() => select(item)} onDoubleClick={() => execute(item)}>
            <img width="24px" height="24px" style={iconStyle} src={item.IconUrl}/>
            <div style={nameStyle}>{item.Name}</div>
            <div style={commentStyle}>{item.Comment}</div>
        </div>
    )
}

export {Item}
