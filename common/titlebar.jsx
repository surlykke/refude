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
        background: "linear-gradient(to bottom, darkgray, lightgray, darkgray)"
    };

    let closeStyle = {
        WebkitAppRegion: "no-drag",
        float: "right",
        width: "calc(1.3em - 4px)",
        height: "calc(1.3em - 4px)",
        background: "linear-gradient(to bottom, lightgray, white, lightgray)",
        padding: "0px",
        margin: "2px"
    };

    return (
        <div style={titleBarStyle}>
            <div style={closeStyle} onClick={() => gui.App.quit()}>
                <svg viewBox="0 0 100 100">
                    <g fillOpacity="0" strokeWidth="10" stroke="black">
                        <rect x="2" y="2" width="96" height="96" strokeWidth="4"/>
                        <line x1="13" y1="13" x2="87" y2="87"/>
                        <line x1="13" y1="87" x2="97" y2="13"/>
                    </g>
                </svg>
            </div>
        </div>
    );

};