/*
 * Copyright (c) Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

import {doPost} from '../common/utils.js'
import { setNavigation, onKeyDown } from "./navigation.js"
import {input, frag } from "../common/elements.js"
import {resourceHead} from './resource.js'
import {linkDivs} from './linkdiv.js'

const startUrl = "/search/desktop"

export class Browser extends React.Component {
    
    constructor(props) {
        super(props)
        this.history = []
        this.resourceUrl = startUrl 
        this.state = { term: "", itemlist: []}
        setNavigation({goTo: this.goTo, goBack: this.goBack, dismiss: this.dismiss})
        this.watchSse()
    }

    goTo = href => {
        this.history.unshift(this.resourceUrl)
        this.resourceUrl = href
        this.setState({ term: "" }, this.getResource)
    }

    goBack = () => {
        this.resourceUrl = this.history.shift() || this.resourceUrl
        this.setState({ term: "" }, this.getResource)
    }

    dismiss = () => {
        this.history = []
        this.resourceUrl = startUrl 
        this.setState({ term: ""})
        doPost("http://localhost:7938/refude/browser/dismiss")
    }

    componentDidMount = () => {
        document.addEventListener("keydown", onKeyDown)
    };

    watchSse = () => {
        let evtSource = new EventSource("http://localhost:7938/watch")

        evtSource.onopen = () => {
            this.getResource()
        }

        evtSource.onerror = event => {
            ipcRenderer.send("sseerror")
            if (evtSource.readyState === 2) {
                setTimeout(watchSse, 5000)
            }
        }

        evtSource.onmessage = event => {
            if (this.resourceUrl === event.data) {
                this.getResource()
            } 
        }
    }

    getResource = () => {
        let href = `${this.resourceUrl}?term=${this.state.term}`
        console.log("Fetching", href)
        fetch(href)
            .then(resp => resp.json())
            .then(
                json => {this.setState({resource: json})},
                error => {this.setState({ resource: undefined })}
            )
    }

    handleInputFocus = e => {
        this.getResource()
    } 

    handleInput = e => {
        this.setState({term: e.target.value}, this.getResource)
    }

    render = () => {
        let { term, resource} = this.state
        return resource ? frag(
            resourceHead(resource),
            input({
                type: 'text',
                className:'search-box', 
                value: term,
                onFocus: () => this.handleInputFocus,
                onInput: this.handleInput, 
                autoFocus: true}),
            linkDivs(resource, this.activate)
        ) : null
    }
}

ReactDOM.render(React.createElement(Browser), document.getElementById('app'))
