// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { div, span, img } from "../common/elements.js"
import { doPost, retrieveResource, follow, followResource } from "../common/utils.js"

export class Main extends React.Component {

    constructor(props) {
        super(props)
        this.state = { }
        followResource("/notification/flash", this.setFlash, this.clearFlash)
    }

    setFlash = flash => {
        console.log("setting flash to:", flash)
        this.setState({flash: flash})
    }

    clearFlash = () => {
        this.setState({flash: undefined})
    }

    render = () => {
        let { flash } = this.state
        if (flash) {
            let size = 48
            if (flash?.data.IconSize > 48) {
                size = Math.min(flash.data.IconSize, 256)
            }
            return div({}, div({ className: "flash" },
                div({ className: "flash-icon" },
                    img({ height: `${size}px`, src: flash.icon, alt: "" })
                ),
                div({ className: "flash-message" },
                    div({ className: "flash-title" }, flash.title),
                    div({ className: "flash-body" }, flash.comment)
                )
            ))
        } else {
            return undefined 
        }
    }
}

let resizeToContent = () => {
    let { width, height } = document.getElementById('main').getBoundingClientRect()
    width = Math.round(window.devicePixelRatio * width)
    height = Math.round(window.devicePixelRatio * height)
    // Java script call window.resizeTo will not make height or width smaller than 50 px (or so),
    // so we ask server to resize us
    doPost("/refude/html/resizeNotifier", { width: width, height: height })
}
new ResizeObserver((observed) => observed && observed[0] && resizeToContent()).observe(document.getElementById('main'))

ReactDOM.createRoot(document.getElementById('main')).render(React.createElement(Main))
