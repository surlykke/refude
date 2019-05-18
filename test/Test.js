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
        Axios.get('http://localhost:7938/device/DisplayDevice').then(resp => {
            let etag = resp.headers.etag
            console.log("etag:" + etag)
            let headers = {"If-None-Match": etag}
            for (let i = 0; i < 30; i++) {
                console.log("Getting...")
                Axios.get('http://localhost:7938/device/DisplayDevice?longpoll', {headers: headers}).then(resp => {
                    console.log("Got", i)
                })
            }
        });
    }

    render = () => {
        return <div>{JSON.stringify(this.state.dd)}</div>
    }
}
