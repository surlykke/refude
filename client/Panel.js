/*
 * Copyright (c) Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

import {div, materialIcon, getJson, doPost, doDelete} from './utils.js'
import {resourceHead} from './Resource.js'
import {links} from './Link.js'

const startUrl = "/search/desktop"

class Panel extends React.Component {
    
    constructor(props) {
        super(props)
        this.history = []
        this.resourceUrl = startUrl 
        this.state = { term: "", itemlist: [] }
        this.watchSse()
    }

    watchSse = () => {
        let evtSource = new EventSource("http://localhost:7938/watch")

        evtSource.onopen = this.getResource

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

    componentDidMount = () => {
       document.addEventListener("keydown", this.onKeyDown)
    };

    getResource = () => {
        console.log("this.resourceUrl", this.resourceUrl)
        let href = `${this.resourceUrl}?term=${this.state.term}`
        fetch(href)
            .then(resp => resp.json())
            .then(
                json => {console.log("json:", json); this.setState({resource: json})},
                error => {console.log(error); this.setState({ resource: undefined })}
            )
    }

    select = link => {
        this.currentLink = link
    }

    activate = (link, dismiss) => {
        if (link.rel === "org.refude.defaultaction" || link.rel === "org.refude.action") {
            doPost(link.href).then(() => dismiss && this.dismiss())
        } else if (link.rel === "org.refude.delete") {
            doDelete(link.href).delete(path).then(() => dismiss && this.dismiss())
        } else if (link.rel === "related") {
            getJson(link.href)
                .then(json => {
                    if (json?._links) {
                        let dfAct = json._links.find(l => l.rel === "org.refude.defaultaction")
                        if (dfAct) {
                            doPost(dfAct.href).then(() => dismiss && this.dismiss())
                        }
                    }
                })
        }
    }

    delete = (link, dismiss) => {
        if (link.rel === "org.refude.delete") {
            doDelete(link.href).delete(path).then(() => dismiss && this.dismiss())
        } else if (link.rel === "related") {
            getJson(link.href)
                .then(json => {
                    if (json?._links) {
                        let deleteLink = json._links.find(l => l.rel === "org.refude.delete")
                        if (deleteLink) {
                            doDelete(deleteLink.href).then(() => dismiss && this.dismiss())
                        }
                    }
                })
        }
    }

    dismiss = () => {
        this.resourceUrl = startUrl 
        this.setState({ term: "" }, this.getResource)
    }

    navigateTo = href => {
        this.history.unshift(this.resourceUrl)
        this.resourceUrl = href
        this.setState({ term: "" }, this.getResource)
    }

    navigateBack = () => {
        this.resourceUrl = this.history.shift() || this.resourceUrl
        this.setState({ term: "" }, this.getResource)
    }

    onKeyDown = (event) => {
        let { key, keyCode, ctrlKey, altKey, shiftKey, metaKey } = event;
        if (key === "ArrowRight" || key === "l" && ctrlKey) {
            this.currentLink.rel === "related" && this.navigateTo(this.currentLink.href)
        } else if (key === "ArrowLeft" || key === "h" && ctrlKey) {
            this.navigateBack()
        } else if (key === "ArrowUp" || key === "k" && ctrlKey || key === 'Tab' && shiftKey && !ctrlKey && !altKey) {
            this.move(true)
        } else if (key === "ArrowDown" || key === "j" && ctrlKey || key === 'Tab' && !shiftKey && !ctrlKey && !altKey) {
            this.move()
        } else if (key === "Enter") {
            this.activate(this.currentLink, !ctrlKey)
        } else if (key === "Delete") {
            this.delete(this.currentLink, !ctrlKey)
        } else if (key === "Escape") {
            this.dismiss()
        } else if (keyCode === 8) {
            this.setState({ term: this.state.term.slice(0, -1) }, this.getResource)
        } else if (key.length === 1 && !ctrlKey && !altKey && !metaKey) {
            this.setState({ term: this.state.term + key }, this.getResource)
        } else {
            return
        }
        event.preventDefault();
    }

    move = up => {
        let items = document.getElementsByClassName("item")
        let idx = Array.from(items).indexOf(document.activeElement)
        if (idx > -1) {
            up ?
                items[(idx + items.length - 1) % items.length].focus() :
                items[(idx + 1) % items.length].focus()
        }
    }


    render = () => {
        let { resource, term } = this.state
        console.log("Panel render..., resource:", resource)
        if (resource) {
            return (
                div({className:"refude", id:"content"},
                    resourceHead(resource),
                    term ? div({className:"searchbox"}, materialIcon("search"), term): null,
                    links(resource, this.activate, this.select)
                )
            ) 
        } else {
            return null
        }
    }
}

ReactDOM.render(React.createElement(Panel), document.getElementById('app'))
