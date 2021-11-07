/*
 * Copyright (c) Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

import { div, frag} from "../common/elements.js"
import { clock } from './clock.js'
import { notifierItem } from './notifieritem.js'
import { battery } from './battery.js'
import { menu } from "./menu.js"
import { doPost } from "../common/utils.js"
import { flash } from "./flash.js"


export class Panel extends React.Component {
    
    constructor(props) {
        super(props)
        this.state = { itemlist: []}
        this.watchSse()
    }

    componentDidMount = () => {
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
            if ("/device/DisplayDevice" === event.data) {
                this.getDisplayDevice()
            } else if ("/item/list" === event.data) {
                this.getItemlist()
            } else if ("/notification/flash" === event.data) {
                this.getFlash()
            } 
        }
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

    getItemlist = () => {
        fetch("http://localhost:7938/item/list")
            .then(resp => resp.json())
            .then( 
                json => {this.setState({itemlist: json.data})},
                error => {this.setState({itemlist: []})}
            )
    }

    setMenuObject = menuObject => this.setState({menuObject: menuObject})

    render = () => {
        let {itemlist, displayDevice, menuObject, flashNotification} = this.state
        return frag(
            div(
                { className: "panel", onClick: () => this.setMenuObject()}, 
                clock(),
                itemlist.map(item => { return notifierItem(item, this.setMenuObject)}),
                battery(displayDevice)
            ),
            menuObject ? menu(menuObject, () => this.setState({menuObject: undefined})) :
            flashNotification ? flash(flashNotification):
            null
        )
    }
}

ReactDOM.render(React.createElement(Panel), document.getElementById('panel'))

let resizeToContent = div => {
   let {width, height} =  div.getBoundingClientRect()

    width = Math.round((width)*window.devicePixelRatio)
    height = Math.round((height)*window.devicePixelRatio)
    console.log("width, height:", width, height)
    doPost("/refude/panel/resize", {width: width, height: height}) 
}

new ResizeObserver((observed) => {
    if (observed && observed[0]) { // shouldn't be nessecary 
        resizeToContent(observed[0].target)    
    }
}).observe(document.getElementById('panel'))

setTimeout(() => resizeToContent(document.getElementById('panel')), 3000)
