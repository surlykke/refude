//Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.

import React from 'react'
import ReactDOM from 'react-dom'
import { ResourceList } from './resourcelist';
import { Resource } from './resource';
import { postUrl, deleteUrl } from '../common/monitor';
import {ipcRenderer} from 'electron'

export class Do extends React.Component {
    constructor(props) {
        super(props);
        this.state = {}
        this.savedTerm = ""
    };

    componentDidMount = () => {
        console.log("Do did mount")
    }

    showRes = (res, term) => {
        this.setState({res: res})
        this.savedTerm = term
    }
    unShowRes = () => this.setState({res: undefined})
    post = path =>  postUrl(path, this.dismiss)
    del = path =>  deleteUrl(path, this.dismiss)
    dismiss = () => {
        this.unShowRes()
        ipcRenderer.send("doClose")
    }

    render = () => {
        return this.state.res ?
            <Resource res={this.state.res} post={this.post} del={this.del} dismiss={this.dismiss} back={this.unShowRes}/> :
            <ResourceList term={this.savedTerm} post={this.post} del={this.del} dismiss={this.dismiss} showRes={this.showRes}/>
    }
}

ReactDOM.render(<Do />, document.getElementById('do'))