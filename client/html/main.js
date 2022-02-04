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
import { activateSelected, deleteSelected, getSelectedAnchor, move, selectDiv, preferred, selectPreferred, setPreferred } from "./navigation.js"
import { resourceHead } from "./resourcehead.js"
import { linkDiv } from "./linkdiv.js"
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
        selectPreferred()
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
        console.log("getFlash")
        fetch("http://localhost:7938/notification/flash")
            .then(resp => resp.json())
            .then(json => {console.log('getFlash got:', json); this.setState({flashNotification: json})},
                  error => this.setState({flashNotification: undefined}))
    }


    getResource = () => {
        console.log("getResource")
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
        console.log("getLinks")
        let browserUrl = this.browserUrl
        if (this.state.resource) {
            let linksUrl = this.state.resource.links
            if (this.state.resource.searchable && this.state.term) {
                let separator = linksUrl.indexOf('?') == -1 ? '?' : '&'
                linksUrl = linksUrl + separator + "search=" + encodeURIComponent(this.state.term)
            }
            if (linksUrl) {
                console.log("fetching", linksUrl)
                fetch(linksUrl)
                    .then(resp => resp.json())
                    .then(json => {
                        console.log("getLinks retrieved:", json) 
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
            move()
        } else {
            this.browserUrl = browserStartUrl
            this.getResource()
        }
    }

    closeBrowser = () => {
        this.browserUrl = undefined
        setPreferred(undefined)
        this.browserHistory = []
        this.setState({term: "", resource: undefined, links: undefined})
    }

    goTo = href => {
        this.browserHistory.unshift({url: this.browserUrl, term: this.state.term, oldPreferred: preferred})
        this.browserUrl = href
        this.setState({ term: "" }, this.getResource)
    }

    goBack = () => {
        let {url, term, oldPreferred} = this.browserHistory.shift()
        this.browserUrl = url || this.browserStartUrl
        console.log("Back, oldPreferred:", oldPreferred)
        setPreferred(oldPreferred)
        this.setState({term: term || ""}, this.getResource)
    }

    handleInput = e => {
        console.log("handleInput", e?.target?.value)
        this.setState({term: e.target.value}, this.getResource)
    }

    onKeyDown = (event) => {
        let { key, ctrlKey, altKey, shiftKey } = event;
        console.log(key, ctrlKey, altKey, shiftKey)
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
            activateSelected(!ctrlKey && this.closeBrowser);
        } else if (key === "Delete") {
            deleteSelected(!ctrlKey && this.closeBrowser);
        } else if (key === "Escape") {
            this.closeBrowser();
            this.setMenuObject();
            this.setState({flashNotification: undefined})
        } else {
            return;
        }

        event.preventDefault();
    }

    linkDblClick = e => {
        selectDiv(e.currentTarget)
        activateSelected(this.closeBrowser)
    }


    render = () => {
        let {itemlist, displayDevice, resource, term, links, menuObject, flashNotification} = this.state
        let elmts = [
            div(
                { className: "panel", onClick: () => this.setMenuObject()}, 
                clock(),
                itemlist.map(item => { return notifierItem(item, this.setMenuObject)}),
                battery(displayDevice)
            )
        ]
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
                        return linkDiv(l, this.linkDblClick)
                    })
                )
            }

            if (links && (resource.searchable || links.length > 0)) {
                if (actionLinks.length > 0) {
                    elmts.push(p({className: "linkHeading related"}, "Related"))
                } 
                if (resource.searchable) {
                    elmts.push(
                        input({
                            type: 'text',
                            className:'search-box', 
                            value: term,
                            onInput: this.handleInput, 
                            autoFocus: true}
                        )
                    )
                }
                elmts.push(...links.map(l => linkDiv(l, this.linkDblClick)))
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
