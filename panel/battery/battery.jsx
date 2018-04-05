// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
import React from 'react'
import {doGet} from '../../common/utils'

const errorData = {
    style: {
        color: 'red',
        fontWeight: 'normal',
        marginRight: '0.4em'
    },
    percentage: '?'
}

class Battery extends React.Component {

    constructor(props) {
        super(props)
        this.onUpdated = props.onUpdated
        this.state = {data: errorData}
    }

    componentDidMount = () => {
        let update = () => {
            let p = doGet("power-service", "/devices/DisplayDevice").then(battery => {
                console.log("Got battery:", battery);

                let charging = ["Charging", "Fully charged"].includes(battery.State)
                this.setState({
                    data: {
                        style: {
                            color: (charging || battery.Percentage >= 20 ? 'black' : 'red'),
                            fontWeight: charging ? 'bold' : 'normal',
                            marginRight: '0.4em'
                        },
                        percentage: Math.floor(battery.Percentage + 0.5),
                    }
                });
            }).catch(err => {console.log("err:", err); this.setState({data: errorData})}).then(setTimeout(update, 1000));
        };
        update();
    };

    componentDidUpdate() {
        this.onUpdated();
    }

    render = () => {
        let d = this.state.data;
        return <div style={this.props.style}><span style={d.style} key={d.path}>{d.percentage}%</span></div>
    }
}

export {
    Battery
}
