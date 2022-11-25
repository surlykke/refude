// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { div, img} from "../common/elements.js"
import { retrieveCollection, restorePosition, savePositionAndClose } from "../common/utils.js"

export class Main extends React.Component {

    constructor(props) {
        super(props)
        this.state = {}
        this.getFlash()
    }

    componentDidMount = () => {
        restorePosition("notify")

        document.addEventListener("dblclick", () => {
            this.pinned = !this.pinned
            console.log("pinned:", this.pinned)
        })
    };

    

    getFlash = () => {
        if (!this.pinned) {
            retrieveCollection("/notification/", this.updateFlash)
        }
        setTimeout(this.getFlash, 500)
    }

    updateFlash = notifications => {
        console.log("updateFlash, notifications:", notifications)
        notifications.reverse() 
        let now = Date.now()
        let notification = notifications.find(n => n.data.Urgency === 2) || // Critical
                                notifications.find(n => n.data.Urgency === 1 && n.data.Created + 20000 > now && n.data.Expires > now) || // Normal
                                notifications.find(n => n.data.Urgency === 0 && n.data.Created + 2000 > now);
        
        if (notification) {
            this.setState({notification: notification})
        } else {
            savePositionAndClose("notify")
        }
    }

    render = () => {
        let {notification } = this.state
        if (notification) {
            let size = 48
            if (notification?.data.IconSize > 48) {
                size = Math.min(notification.data.IconSize, 256)
            }
            return (
                div({ className: "flash" },
                    div({ className: "flash-icon" },
                        img({ height: `${size}px`, src: notification.icon, alt: "" })
                    ),
                    div({ className: "flash-message" },
                        div({ className: "flash-title" }, notification.title),
                        div({ className: "flash-body" }, notification.comment)
                    )
                )
            )

        } else {
            return div({}) 
        }
    }
}

let resizeToContent = () => {
    let { width, height } = document.getElementById('main').getBoundingClientRect()
    width = Math.round(window.devicePixelRatio*width)
    height = Math.round(window.devicePixelRatio*height)

    window.resizeTo(Math.max(width, 180), Math.max(40, height))
}
new ResizeObserver((observed) => observed && observed[0] && resizeToContent()).observe(document.getElementById('main'))

ReactDOM.createRoot(document.getElementById('main')).render(React.createElement(Main))
