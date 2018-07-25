// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project. 
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';

export class SearchBox extends React.Component {

    constructor(props) {
        super(props);
        this.inputField = React.createRef();
    }

    onChange = (event) => {
        this.props.onChange(event.target.value)
    };

    clear = () => {
        this.inputField.current.value = "";
    };

    render = () => {
        let style = {
            boxSizing: "border-box",
            paddingRight: "5px",
        };

        Object.assign(style, this.props.style);

        let inputStyle = {
            width: "100%",
            height: "36px",
            borderRadius: "5px",
            outlineStyle: "none",
        };

        return (
            <div style={style}>
                <input style={inputStyle} type="search" onChange={this.onChange} ref={this.inputField} autoFocus/>
            </div>
        )
    }
}

