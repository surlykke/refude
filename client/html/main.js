// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { div, frag, input, p} from "./elements.js"
import { clock } from './clock.js'
import { notifierItem } from './notifieritem.js'
import { battery } from './battery.js'
import { doPost } from "./utils.js"
import { flash } from "./flash.js"
import { resourceHead } from "./resourcehead.js"
import { link } from "./link.js"
import { menu } from "./menu.js"

const browserStartUrl = "/start"


export class Main extends React.Component {
    
    constructor(props) {
        super(props)
        this.state = { itemlist: [], term: ""}
        this.browserHistory = []
        this.watchSse()
    }

    componentDidMount = () => {
        document.addEventListener("keydown", this.onKeyDown)
        document.getElementById('panel').addEventListener('focusout', () => this.setState({menuObject: undefined}))
    };

    componentDidUpdate = () => {
        let links = Array.from(document.querySelectorAll("a.link"))
        let link = links.find(l => l.href == this.preferred) || links[0]
        link?.focus()
    }

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
            if ("/item/list" === event.data) {
                this.getItemlist()
            } else if ("/device/DisplayDevice" === event.data) {
                this.getDisplayDevice()
            } else if ("/notification/flash" === event.data) {
                this.getFlash()
            } else if ("/refude/openBrowser" === event.data) {
                this.openBrowser()
            } else if (this.browserUrl && this.browserUrl === event.data) {
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
        let browserUrl = this.browserUrl
        fetch(browserUrl)
            .then(resp => resp.json())
            .then(
                json => {
                    // browserUrl may have changed while request in flight 
                    if (browserUrl === this.browserUrl) {
                        this.setState({resource: json}, this.getLinks)
                    }
                }, 
                error => {
                    console.log("getResource error:", error)
                    this.setState({ resource: undefined })
                }
            )
    }

    getLinks = () => {
        let browserUrl = this.browserUrl
        if (this.state.resource) {
            let linksUrl = this.state.resource.links
            if (this.state.resource.searchable && this.state.term) {
                let separator = linksUrl.indexOf('?') == -1 ? '?' : '&'
                linksUrl = linksUrl + separator + "search=" + encodeURIComponent(this.state.term)
            }
            if (linksUrl) {
                fetch(linksUrl)
                    .then(resp => resp.json())
                    .then(json => {
                        if (browserUrl === this.browserUrl) {
                            this.setState({links: json})
                        }
                    }, 
                    error => {
                        console.log("error:", error)
                        this.setState({links: undefined})
                    })
            } else {
                this.setState({links: undefined})
            }
        }
    }

    setMenuObject = menuObject => this.setState({menuObject: menuObject})

    openBrowser = () => {
        if (this.state.links) {
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
        this.setState({term: "", resource: undefined, links: undefined})
    }

    handleInput = e => {
        this.setState({term: e.target.value}, this.getResource)
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
                this.browserHistory.unshift({url: this.browserUrl, term: this.state.term, oldPreferred: this.preferred})
                this.browserUrl = href
                this.setState({ term: "" }, this.getResource)
            }
        } else { // left
            let {url, term, oldPreferred} = this.browserHistory.shift()
            this.browserUrl = url || this.browserStartUrl
            this.preferred = oldPreferred
            this.setState({term: term || ""}, this.getResource)
        }
    }

    onKeyDown = (event) => {
        let { key, ctrlKey, altKey } = event;
        console.log("key:", key)
        if (key === "Escape") {
            this.closeBrowser();
            this.setMenuObject();
            this.setState({flashNotification: undefined})
        } else if (key.length === 1 && !ctrlKey && !altKey) {
            this.setState({term: this.state.term + key}, this.getResource)            
        } else if (key === "Backspace") {
            this.setState({term: this.state.term.slice(0, -1)}, this.getResource)
        } else {
            return 
        }
        event.preventDefault();
    }

    render = () => {
        let {itemlist, displayDevice, resource, term, links, menuObject, flashNotification} = this.state
        let elmts = [
            div(
                { className: "panel", onClick: () => this.setMenuObject()}, 
                clock(),
                itemlist.map(item => { return notifierItem(item, this.setMenuObject)}),
                battery(displayDevice)
            )        ]
        if (resource) {
            elmts.push(resourceHead(resource))
            let actionLinks = [...resource.post]
            if (resource.delete) {
                actionLinks.push(resource.delete)
            }
            if (actionLinks.length > 0) {
                elmts.push(
                    p({className: "linkHeading"}, "Actions"),
                    ...actionLinks.map(l => {
                        return link(l, "Action", this.closeBrowser, this.move)
                    })
                )
            }

            if (links && (resource.searchable || links.length > 0)) {
                if (actionLinks.length > 0) {
                    elmts.push(p({className: "linkHeading related"}, "Related"))
                } 
                elmts.push(
                    div({className:'search-box'}, term)
                )
                elmts.push(...links.map(l => link(l, l.profile, this.closeBrowser, this.move)))
            }
        } else if (menuObject) {
            elmts.push(menu(menuObject, () => this.setState({menuObject: undefined})))
        } else if (flashNotification) {
            elmts.push(flash(flashNotification))
        }
       
        return frag( ...elmts)
    }
}

ReactDOM.render(React.createElement(Main), document.getElementById('panel'))

let resizeToContent = div => {
   let {width, height} =  div.getBoundingClientRect()

    width = Math.round((width)*window.devicePixelRatio)
    height = Math.round((height)*window.devicePixelRatio)
    doPost("/refude/resizePanel", {width: width, height: height}) 
}

new ResizeObserver((observed) => {
    if (observed && observed[0]) { // shouldn't be nessecary 
        resizeToContent(observed[0].target)    
    }
}).observe(document.getElementById('panel'))

setTimeout(() => resizeToContent(document.getElementById('panel')), 3000)
