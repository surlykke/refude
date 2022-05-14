// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { div, frag, input, p, span } from "./elements.js"
import { clock } from './clock.js'
import { notifierItem } from './notifieritem.js'
import { battery } from './battery.js'
import { doPost, followResource, watchResource } from "./utils.js"
import { flash } from "./flash.js"
import { resourceHead } from "./resourcehead.js"
import { link } from "./link.js"
import { menu } from "./menu.js"

const browserStartUrl = "http://localhost:7938/start"


export class Main extends React.Component {

    constructor(props) {
        super(props)
        this.state = { itemlist: [], term: "" }
        this.browserHistory = []
        followResource("/notification/", this.updateFlash, () => this.setState({flash: undefined}))
        followResource("/item/", itemlist => this.setState({itemlist: itemlist}), () => this.setState({itemlist: []}))
        followResource("/device/", 
                       deviceList => this.setState({displayDevice: deviceList.find(d => d.data.DisplayDevice)?.data}), 
                       () => this.setState({displayDevice: undefined}))
        watchResource("/refude/openBrowser", this.openBrowser)
    }

    componentDidMount = () => {
        document.addEventListener("keydown", this.onKeyDown)
        document.getElementById('panel').addEventListener('focusout', () => this.setState({ menuObject: undefined }))
    };

    componentDidUpdate = () => {
        let links = Array.from(document.querySelectorAll("a.link"))
        let link = links.find(l => l.href == this.preferred) || links[0]
        link?.focus()
    }

    updateFlash = notifications => {
        notifications.reverse() 
        let now = Date.now()
        let flashNotification = notifications.find(n => n.data.Urgency === 2) || // Critical
                                notifications.find(n => n.data.Urgency === 1 && n.data.Created + 6000 > now) || // Normal
                                notifications.find(n => n.data.Urgency === 0 && n.data.Created + 2000 > now);
        
        console.log("flashNotification:", flashNotification)


        if (flashNotification) {
            this.setState({flashNotification: flashNotification})
            if (flashNotification.data.Urgency < 2) {
                let timeout = flashNotification.data.Urgency === 1 ? 6050 : 2050
                console.log("Scheduling this.removeFlash", timeout, this.removeFlash)
                setTimeout(this.removeFlash, timeout)
            }
        }
    }

    removeFlash = () => {
        console.log("removeFlash")
        let {flashNotification: fn} = this.state
        if (fn && fn.data.Urgency < 2) {
            if (fn.data.Created + (fn.data.Urgency == 1 ? 6000 : 2000) < Date.now()) {
                this.setState({flashNotification: undefined})
            }
        }
    }

    getResource = () => {
        if (!this.browserUrl) {
            return
        }
        let browserUrlCopy = this.browserUrl
        let dummy = new URL("http://localhost:7938?foo=FOO")
        console.log("dummy:", dummy.toString())
        dummy.searchParams.append("baa", "BAA")
        console.log("dummy:", dummy.toString())
        
        console.log("this.browserUrl", this.browserUrl)
        let url = new URL(this.browserUrl)
        url.searchParams.append("search", this.state.term)
        console.log("fetching ", url.toString())
        fetch(url)
            .then(resp => resp.json())
            .then(
                json => {
                    // browserUrl may have changed while request in flight 
                    if (browserUrlCopy === this.browserUrl) {
                        this.setState({ resource: json })
                    }
                },
                error => {
                    console.log("getResource error:", error)
                    this.setState({ resource: undefined })
                }
            )
    }

    setMenuObject = menuObject => this.setState({ menuObject: menuObject })

    openBrowser = () => {
        if (this.state.resource) {
            this.move("down")
        } else {
            this.browserUrl = browserStartUrl
            this.getResource()
        }
    }

    closeBrowser = () => {
        this.browserUrl = undefined
        this.preferred = undefined
        this.browserHistory = []
        this.setState({ term: "", resource: undefined, links: undefined })
    }

    handleInput = e => {
        this.setState({ term: e.target.value }, this.getResource)
    }

    move = direction => {
        if (direction === "up" || direction === "down") {
            let list = Array.from(document.querySelectorAll("a.link"))
            let currentIndex = list.indexOf(document.activeElement)
            if (list.length < 2 || currentIndex < 0) {
                list[0]?.focus();
            } else {
                list[(currentIndex + list.length + (direction === "up" ? -1 : 1)) % list.length].focus()
                this.preferred = document.activeElement.href
            }
        } else if (direction === "right") {
            let href = document.activeElement.href
            if (href) {
                this.browserHistory.unshift({ url: this.browserUrl, term: this.state.term, oldPreferred: this.preferred })
                this.browserUrl = href
                this.setState({ term: "" }, this.getResource)
            }
        } else { // left
            let { url, term, oldPreferred } = this.browserHistory.shift() || {}
            this.browserUrl = url || browserStartUrl
            this.preferred = oldPreferred
            this.setState({ term: term || "" }, this.getResource)
        }
    }

    onKeyDown = (event) => {
        let { key, ctrlKey, altKey } = event;
        if (key === "Escape") {
            this.closeBrowser();
            this.setMenuObject();
            this.setState({ flashNotification: undefined })
        } else if (key === "ArrowLeft" || key === "h" && ctrlKey) {
            this.move("left")
        } else if (key.length === 1 && !ctrlKey && !altKey) {
            this.setState({ term: this.state.term + key }, this.getResource)
        } else if (key === "Backspace") {
            this.setState({ term: this.state.term.slice(0, -1) }, this.getResource)
        } else {
            return
        }
        event.preventDefault();
    }

    render = () => {
        let { itemlist, displayDevice, resource, term, menuObject, flashNotification } = this.state
        console.log("render, displayDevice:", displayDevice)
        return frag(
            div(
                { className: "panel", onClick: () => this.setMenuObject() },
                clock(),
                itemlist.map(item => { return notifierItem(item, this.setMenuObject) }),
                battery(displayDevice)
            ),
            resource ? frag(
                resourceHead(resource),
                div({ className: 'search-box' },
                    span({ style: { display: term ? "" : "none" } }, term)
                ),
                term && resource.links.length === 0 && div({className: 'linkHeading'}, "No match"),
                div({ className: 'links' }, ...resource.links.map(l => link(l, l.profile, this.closeBrowser, this.move)))
            ) : menuObject ? menu(menuObject, () => this.setState({ menuObject: undefined }))
                : flashNotification ? flash(flashNotification)
                    : null
        )
    }
}

ReactDOM.render(React.createElement(Main), document.getElementById('panel'))

let resizeToContent = div => {
    let { width, height } = div.getBoundingClientRect()

    width = Math.round((width) * window.devicePixelRatio)
    height = Math.round((height) * window.devicePixelRatio)
    doPost("/refude/resizePanel", { width: width, height: height })
}

new ResizeObserver((observed) => {
    if (observed && observed[0]) { // shouldn't be nessecary 
        resizeToContent(observed[0].target)
    }
}).observe(document.getElementById('panel'))

setTimeout(() => resizeToContent(document.getElementById('panel')), 3000)
