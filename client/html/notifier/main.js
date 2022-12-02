// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { div, img} from "../common/elements.js"
import { retrieveCollection, restorePosition, savePositionAndClose, doPost, followCollection, follow } from "../common/utils.js"

const lowDuration = 4000
const normalDuration = 10000 

export class Main extends React.Component {

    constructor(props) {
        super(props)
        this.state = {}
        followCollection("/notification/", this.updateFlash, this.errorHandler)
    }

    componentDidMount = () => {
        restorePosition("notify")
        setTimeout(resizeToContent, 750)
    };

    updateFlash = () => {
        console.log("Into updateFlash")
        retrieveCollection(
            "/notification/",
            notifications => {
                notifications.reverse() 
                let now = Date.now()
                let notification = notifications.find(n => n.data.Urgency === 2) || // Critical
                                        notifications.find(n => n.data.Urgency === 1 && n.data.Created + normalDuration > now) || // Normal
                                        notifications.find(n => n.data.Urgency === 0 && n.data.Created + lowDuration > now);

                console.log("notification:", notification)
                console.log("now:", now)
                console.log("Created: ", notification.data.Created, " - ", notification.data.Created - now)
                console.log("Expires: ", notification.data.Expires, " - ", notification.data.Expires - now)
                this.setState({notification: notification})
                
                if (notification?.data?.Urgency === 0) {
                    this.deadline = notification.data.Created + lowDuration
                } else if (notification?.data?.Urgency == 1) { 
                    this.deadline = notification.data.Created + normalDuration
                } else {
                    this.deadline = notification.data.Expires     
                }
                console.log("deadline:", this.deadline, " - ", this.deadline - now)
                if (this.deadline){
                    let timeout =  Math.max(10, this.deadline - now)
                    console.log("Timeout:", timeout)
                    setTimeout(this.updateFlash, timeout)
                }
            },
            this.errorHandler
        )
    }

    errorHandler = e => this.setState({notification: undefined})


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
            return div({className: "placeholder"}, "-") 
        }
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
