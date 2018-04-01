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
import {adjustIconUrl, doGet} from '../../common/utils'

class Battery extends React.Component {

    constructor(props) {
        super(props)
        this.onUpdated = props.onUpdated
        this.state = {data: []}
    }

    componentDidMount = () => {
        let compare = (b1, b2) => b1.NativePath.localeCompare(b2.NativePath);
        let update = () => {
            doGet("power-service", "/search", {type: "battery"}).then(batteries => {
                let data = batteries.sort(compare).map(b => {
                    let charging = ["Charging", "Fully charged"].includes(b.State)
                    return {
                        style: {
                            color: (charging || b.Percentage >= 20 ? 'black' : 'red'),
                            fontWeight: charging ? 'bold' : 'normal',
                            marginRight: '0.4em'
                        },
                        percentage: Math.floor(b.Percentage + 0.5),
                        path: b.NativePath
                    }
                });
                this.setState({data: data});
                setTimeout(update, 1000);
            });
        };

        update();
    };

    componentDidUpdate

    render = () => {
            return <div style={this.props.style}>{this.state.data.map(d => (
                <span style={d.style} key={d.path}>{d.percentage}%</span>))}</div>
        }
    }

    export {
    Battery
}
