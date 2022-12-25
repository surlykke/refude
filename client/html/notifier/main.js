// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { div, frag, img} from "../common/elements.js"
import { doPost, retrieveResource, follow } from "../common/utils.js"

const leftPad = num => num < 10 ? "0" + num : "" + num

export class Main extends React.Component {

    constructor(props) {
        super(props)
        this.notifications = []
        this.state = {time: '...'}
        follow("/notification/", this.updateFlash, this.clearFlash)
        setTimeout(this.keepTime, 2000)
    }

    keepTime = () => {
        let now = new Date()
        let timeStr =
            `${now.getFullYear()}-${leftPad(now.getMonth() + 1)}-${leftPad(now.getDate())} ` +
            `${leftPad(now.getHours())}:${leftPad(now.getMinutes())}:${leftPad(now.getSeconds())}`
        this.setState({time: timeStr})
        setTimeout(this.keepTime, 1010 + 10 - now.getMilliseconds())
    }

    updateFlash = () => {
        retrieveResource("/notification/flash", 
        flash => {
            this.setState({flash: flash})
            let {Created, Urgency} = flash.data
            let expires = Created + (Urgency === "2" ? 3600000 : Urgency === 1 ? 10000 : 4000)
            setTimeout(this.updateFlash, expires - new Date().getTime() + 10)
        }, 
        this.clearFlash)
    }

    clearFlash = () => this.setState({flash: undefined})

    render = () => {
        let {time, flash} = this.state
        let size = 48
        if (flash?.data.IconSize > 48) {
            size = Math.min(flash.data.IconSize, 256)
        }
        return frag(
            div({className: "date"}, time),
            div({}, 
            flash && div({ className: "flash" },
                                div({ className: "flash-icon" },
                                    img({ height: `${size}px`, src: flash.icon, alt: "" })
                                ),
                                div({ className: "flash-message" },
                                    div({ className: "flash-title" }, flash.title),
                                    div({ className: "flash-body" }, flash.comment)
                                )
                            )
            )
        )
    }
}

let resizeToContent = () => {
    let { width, height } = document.getElementById('main').getBoundingClientRect()
    width = Math.round(window.devicePixelRatio*width)
    height = Math.round(window.devicePixelRatio*height)

    // Java script call window.resizeTo will not make height or width smaller than 50 px (or so),
    // so we ask server to resize us
    doPost("/refude/html/resizeNotifier", {width: width, height: height})
}
new ResizeObserver((observed) => observed && observed[0] && resizeToContent()).observe(document.getElementById('main'))

ReactDOM.createRoot(document.getElementById('main')).render(React.createElement(Main))
