// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';


export class PopUp extends React.Component {

    render = () => {
        let overlayStyle = {
            position: "fixed",
            top: "0px",
            left: "0px",
            width: "100%",
            height: "100%",
            zAxis: "10",
            backgroundColor: "rgba(255, 255, 255, 0.7)",
        };

        let popupStyle = {
            position: "absolute",
            top: "3em",
            left: "20%",
            padding: "16px",
            height: "8em",
            width: "calc(60% - 36px)",
            backgroundColor: "rgba(255, 255, 255, 1)",
            borderRadius: "5px",
            border: "solid black 2px"
        };

        return (
            <div style={overlayStyle} onKeyDown={this.props.dismiss}>
                <div style={popupStyle}>
                   {this.props.children}
                </div>
            </div>
        );

    }


}