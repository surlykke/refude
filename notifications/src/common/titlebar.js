// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';

const gui = window.require('nw.gui');
export let TitleBar = props => {
    let quit = () => {
        console.log("Quitting..");
        gui.App.quit();
    };

    let titleBarStyle = {
        WebkitAppRegion: "drag",
        width: "100%",
        height: "1.3em",
        backgroundColor: "lightgray"
    };

    let closeStyle = {
        WebkitAppRegion: "no-drag",
        float: "right",
        width: "calc(1.3em - 4px)",
        height: "calc(1.3em - 4px)",
        padding: "0px",
        margin: "2px"
    };

    return (
        <div style={titleBarStyle}>
            <div style={closeStyle} onClick={() => gui.App.quit()}>
                <svg viewBox="0 0 100 100">
                    <g fillOpacity="0" strokeWidth="8" stroke="white">
                        <line x1="15" y1="15" x2="85" y2="85"/>
                        <line x1="15" y1="85" x2="85" y2="15"/>
                    </g>
                </svg>
            </div>
        </div>
    );

};