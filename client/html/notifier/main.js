// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { div, img } from "../common/elements.js"
import { doPost, retrieveResource, follow} from "../common/utils.js"

export class Main extends React.Component {

    constructor(props) {
        super(props)
        this.state = { }
        follow("/notification/",  this.getFlash, window.close)
    }

    getFlash = () => retrieveResource("/notification/flash", flash => this.setState({flash: flash}), window.close)

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
    // Put it in lower-right corner of whatever screen we are on
    let { width, height } = document.getElementById('main').getBoundingClientRect()
    width = Math.max(width, 100) 
    height = Math.max(height, 100)
    let x = screen.availLeft + screen.availWidth - width - 1;
    let y = screen.availTop + screen.availHeight - height - 1;
    console.log("resizeTo", width, height)
    window.resizeTo(width, height)
    window.moveTo(x, y)
}
new ResizeObserver((observed) => observed && observed[0] && resizeToContent()).observe(document.getElementById('main'))

ReactDOM.createRoot(document.getElementById('main')).render(React.createElement(Main))
