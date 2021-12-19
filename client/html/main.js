/*
 * Copyright (c) Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

import { div, frag, input} from "./elements.js"
import { clock } from './clock.js'
import { notifierItem } from './notifieritem.js'
import { battery } from './battery.js'
import { menu } from "./menu.js"
import { doPost } from "./utils.js"
import { flash } from "./flash.js"
import { activateSelected, deleteSelected, getSelectedAnchor, move } from "./navigation.js"
import { resourceHead } from "./resource.js"
import { linkDivs } from "./linkdiv.js"

const browserStartUrl = "/search/desktop"


export class Main extends React.Component {
    
    constructor(props) {
        super(props)
        this.state = { itemlist: [], term: "", browserHistory: []}
        this.watchSse()
    }

    componentDidMount = () => {
        document.addEventListener("keydown", this.onKeyDown)
        document.getElementById('panel').addEventListener('focusout', () => this.setState({menuObject: undefined}))
    };

    watchSse = () => {
        let evtSource = new EventSource("http://localhost:7938/watch")

        evtSource.onopen = () => {
            this.getDisplayDevice()
            this.getItemlist()
            this.getFlash()
        }

        evtSource.onerror = event => {
            this.setState({itemList: [], displayDevice: undefined})
            if (evtSource.readyState === 2) {
                setTimeout(watchSse, 5000)
            }
        }

        evtSource.onmessage = event => {
            console.log("watchSse, event:", event)
            if ("/item/list" === event.data) {
                this.getItemlist()
            } else if ("/device/DisplayDevice" === event.data) {
                this.getDisplayDevice()
            } else if ("/notification/flash" === event.data) {
                this.getFlash()
            } else if ("/refude/openBrowser" === event.data) {
                this.openBrowser()
            } else if (this.resourceUrl === event.data) {
                this.getResource()
            }
        }
    }

    getItemlist = () => {
        fetch("http://localhost:7938/item/list")
            .then(resp => resp.json())
            .then( 
                json => {this.setState({itemlist: json.data})},
                error => {this.setState({itemlist: []})}
            )
    }


    getDisplayDevice = () => {
        fetch("http://localhost:7938/device/DisplayDevice")
            .then(resp => resp.json())
            .then(
                json => this.setState({displayDevice: json.data}),
                error => this.setState({displayDevice: undefined})
            )
    }

    getFlash = () => {
        fetch("http://localhost:7938/notification/flash")
            .then(resp => resp.json())
            .then(json => this.setState({flashNotification: json}),
                  error => this.setState({flashNotification: undefined}))
    }


    getResource = () => {
        let href = `${this.browserUrl}?term=${this.state.term}`
        console.log("Fetching", href)
        fetch(href)
            .then(resp => resp.json())
            .then(
                json => {this.setState({resource: json})},
                error => {this.setState({ resource: undefined })}
            )
    }


    setMenuObject = menuObject => this.setState({menuObject: menuObject})

    openBrowser = () => {
        console.log("open browser..")
        if (this.browserUrl) {
            move()
        } else {
            this.browserUrl = browserStartUrl
            this.getResource()
        }
    }

    closeBrowser = () => {
        this.browserUrl = undefined
        this.browserHistory = []
        this.setState({term: "", resource: undefined})
    }

    goTo = href => {
        this.browserHistory.unshift(this.browserUrl)
        this.browserUrl = href
        this.setState({ term: "" }, this.getResource)
    }

    goBack = () => {
        this.browserUrl = this.browserHistory.shift() || this.browserUrl
        this.setState({ term: "" }, this.getResource)
    }

    handleInput = e => {
        this.setState({term: e.target.value}, this.getResource)
    }

    onKeyDown = (event) => {
        let { key, ctrlKey, altKey, shiftKey } = event;
        if (key === "ArrowRight" || key === "l" && ctrlKey) {
            let selectedAnchor = getSelectedAnchor();
            if (selectedAnchor?.rel === "related") {
                this.goTo(selectedAnchor.href);
            }
        } else if (key === "ArrowLeft" || key === "h" && ctrlKey) {
            this.goBack();
        } else if (key === "ArrowUp" || key === "k" && ctrlKey || key === 'Tab' && shiftKey && !ctrlKey && !altKey) {
            move(true);
        } else if (key === "ArrowDown" || key === "j" && ctrlKey || key === 'Tab' && !shiftKey && !ctrlKey && !altKey) {
            move();
        } else if (key === "Enter") {
            console.log("Activate");
            activateSelected(!ctrlKey && this.closeBrowser);
        } else if (key === "Delete") {
            deleteSelected(!ctrlKey && this.closeBrowser);
        } else if (key === "Escape") {
            this.closeBrowser();
        } else {
            return;
        }

        event.preventDefault();
    }


    render = () => {
        let {itemlist, displayDevice, resource, term, menuObject, flashNotification} = this.state
        console.log("panel render, resource:", resource)
        return frag(
            div(
                { className: "panel", onClick: () => this.setMenuObject()}, 
                clock(),
                itemlist.map(item => { return notifierItem(item, this.setMenuObject)}),
                battery(displayDevice)
            ),
            resource ? frag(
                resourceHead(resource),
                input({
                    type: 'text',
                    className:'search-box', 
                    value: term,
                    onInput: this.handleInput, 
                    autoFocus: true}),
                linkDivs(resource)
            ) : 
            menuObject ? menu(menuObject, () => this.setState({menuObject: undefined})) :
            flashNotification ? flash(flashNotification):
            null
        )
    }
}

ReactDOM.render(React.createElement(Main), document.getElementById('panel'))

let resizeToContent = div => {
   let {width, height} =  div.getBoundingClientRect()

    width = Math.round((width)*window.devicePixelRatio)
    height = Math.round((height)*window.devicePixelRatio)
    console.log("width, height:", width, height)
    doPost("/refude/resizePanel", {width: width, height: height}) 
}

new ResizeObserver((observed) => {
    if (observed && observed[0]) { // shouldn't be nessecary 
        resizeToContent(observed[0].target)    
    }
}).observe(document.getElementById('panel'))

setTimeout(() => resizeToContent(document.getElementById('panel')), 3000)
