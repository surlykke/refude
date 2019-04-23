// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react';
import Axios from 'axios';
import {devtools} from '../common/utils'

export default class Test extends React.Component {
    constructor(props) {
        super(props);
        this.state = {dd: ""}
        devtools()
    }

    componentDidMount = () => {
        Axios({socketPath: '/run/user/1000/org.refude.desktop-service', url: "http:/localhost/device/DisplayDevice"}).then(resp => {
            this.setState({dd: resp.data})
        });
    }

    render = () => {
        return <div>{JSON.stringify(this.state.dd)}</div>
    }
}
