// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { div, frag, span } from "../common/elements.js"
import { resourceHead } from "./resourcehead.js"
import { link } from "./link.js"
import { doPost, restorePosition } from "../common/utils.js"

const browserStartUrl = "http://localhost:7938/start"


export class Main extends React.Component {

    constructor(props) {
        super(props)
        this.state = { term: "" }
        this.browserUrl = browserStartUrl
        this.browserHistory = []
        this.getResource()
    }

    componentDidMount = () => {
        document.addEventListener("keydown", this.onKeyDown)
    };

    componentDidUpdate = () => {
        let links = Array.from(document.querySelectorAll("a.link"))
        let link = links.find(l => l.href == this.preferred) || links[0]
        link?.focus()
    }

    getResource = () => {
        if (!this.browserUrl) {
            return
        }
        let browserUrlCopy = this.browserUrl
        let url = new URL(this.browserUrl)
        url.searchParams.append("search", this.state.term)
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
                    console.warn("getResource error:", error)
                    this.setState({ resource: undefined })
                }
            )
    }

    closeBrowser = () => {
        doPost("http://localhost:7938/refude/hideLauncher")
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
        let { resource, term } = this.state
        let fraqs = []
        if (resource) {
            fraqs.push(
                resourceHead(resource),
                div({ className: 'search-box' },
                    span({ style: { display: term ? "" : "none" } }, term)
                ) 
            )
            if (term && resource.links.length === 0) {
                fraqs.push(div({className: 'linkHeading'}, "No match"))
            }
            let links = resource.links
                            .filter(l => l.title !== "Refude launcher++")
                            .map(l => link(l, l.profile, this.closeBrowser, this.move))
            let firstRel = links.findIndex(l => l.props.rel === 'related')
            if (firstRel > 0) {
                links[firstRel - 1].props.className +=" last-action"
            }
            fraqs.push(div({ className: 'links' }, ...links))
        }

        return frag(fraqs)
    }
}

let resizeToContent = () => {
    let { width, height } = document.getElementById('main').getBoundingClientRect()
    window.resizeTo(width + 100, Math.min(600, Math.max(300, height)))
}
new ResizeObserver((observed) => observed && observed[0] && resizeToContent()).observe(document.getElementById('main'))

ReactDOM.createRoot(document.getElementById('main')).render(React.createElement(Main))
