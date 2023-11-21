// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { div, frag, hr, span, a } from "../common/elements.js"
import { resourceHead } from "./resourcehead.js"
import { link } from "./link.js"
import { doPost} from "../common/utils.js"

const browserStartUrl = "http://localhost:7938/start"


export class Main extends React.Component {

    constructor(props) {
        super(props)
        this.state = { term: "" }
        this.browserUrl = browserStartUrl
        this.browserHistory = []
        this.getResource()
        this.watchSearch()
    }

    componentDidMount = () => {
        document.addEventListener("keydown", this.onKeyDown)
    };

    componentDidUpdate = () => {
        let links = Array.from(document.querySelectorAll("a.link"))
        let link = links.find(l => l.href == this.preferred) || links[0]
        link?.focus()
    }

    watchSearch = () => {
        let evtSource = new EventSource("http://localhost:7938/watch")
        evtSource.addEventListener("resourceChanged", ({data}) => data === "/start" && this.getResource())
        evtSource.onerror = () => {
            if (evtSource.readyState === 2) {
                setTimeout(this.watchSearch, 5000)
            }
        }
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
                        console.log("Resource:", json)
                        this.setState({ resource: json })
                    }
                },
                error => {
                    console.warn("getResource error:", error)
                    this.setState({ resource: undefined })
                }
            )
    }

    dismiss = (restoreWindow, restoreTab) => {
        let url = "http://localhost:7938/refude/html/hidelauncher"
        let separator = "?"
        if (restoreTab) {
            url = url + separator + "restore=tab"
            separator = "&"
        }
        if (restoreWindow) {
            url = url + separator + "restore=window"
        }
        doPost(url)
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
            this.dismiss(true, true);
        } else if (key === "ArrowLeft") {
            this.move("left")
        } else if (key.length === 1 && !ctrlKey && !altKey) {
            this.setState({ term: this.state.term + key }, this.getResource)
        } else if (key === "Backspace") {
            this.setState({ term: this.state.term.slice(0, -1) }, this.getResource)
        } else if (key === "n" && ctrlKey || key == "p" && ctrlKey) {
            // Ignore
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
                fraqs.push(div({ className: 'linkHeading' }, "No match"))
            }
            let links = resource.links ? resource.links.map(l => link(l, l.profile, this.dismiss, this.move)) : []
            let firstRel = links.findIndex(l => l.props.rel === 'related')
            if (firstRel > 0) {
                links = [span({ className: "linkHeading" }, "Actions")]
                    .concat(...links.slice(0, firstRel))
                    .concat(span({ className: "linkHeading" }, "Links"))
                    .concat(...links.slice(firstRel))
            }
            fraqs.push(div({ className: 'links' }, ...links))
        }

        return frag(fraqs)
    }
}

ReactDOM.createRoot(document.getElementById('main')).render(React.createElement(Main))
