import React from 'react';

export let TitleBar = props => {
    let titleBarStyle = {
        WebkitAppRegion: "drag",
        width:"100%",
        height:"1.3em",
        background: "linear-gradient(to bottom, darkgray, lightgray, darkgray)"
    };
    return  <div style={titleBarStyle}/>;

}