// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { div, frag, img, span} from "../common/elements.js"
import { retrieveCollection, restorePosition, savePositionAndClose, doPost, followCollection, follow } from "../common/utils.js"

const lowDuration = 4000
const normalDuration = 10000 
const criticalDuration = 3600000

export class Main extends React.Component {

    constructor(props) {
        super(props)
        followCollection("/notification/", this.updateFlash, this.errorHandler)
        this.state = {placeholder: ' '}
    }

    componentDidMount = () => {
        restorePosition("notify")
        setTimeout(() => this.setState({placeholder: '◯'}), 1000)
    };

    updateFlash = () => {
        console.log("Into updateFlash")
        retrieveCollection(
            "/notification/",
            notifications => {
                notifications.reverse() 
                // We don't really care about Expires, but show notifications with low urgency for 4 secs, 
                // normal 10 secs, and critical for an hour (or until dismissed)
                // We show the last critical, or if not found, last normal or if not found last low
                let now = Date.now()
                let notification = 
                    notifications.find(n => n.data.Urgency === 2) || // Critical 
                    notifications.find(n => n.data.Urgency === 1 && n.data.Created + normalDuration > now) || // Normal 
                    notifications.find(n => n.data.Urgency === 0 && n.data.Created + lowDuration > now);

                this.setState({notification: notification})
                let urgency = notification?.data?.Urgency 
                let duration = urgency === 0 ? lowDuration : urgency === 1 ? normalDuration : criticalDuration
                let timeout = notification?.data?.Created + duration - now
                if (timeout){
                    setTimeout(this.updateFlash, Math.max(10, timeout))
                }
            },
            this.errorHandler
        )
    }

    errorHandler = e => this.setState({notification: undefined})


    render = () => {
        let {notification, placeholder} = this.state
        let size 
        if (notification) {
            size = 48
            if (notification?.data.IconSize > 48) {
                size = Math.min(notification.data.IconSize, 256)
            }
            return div({}, 
                notification && div({ className: "flash" },
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
            return span({className: 'placeholder'}, placeholder)
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
