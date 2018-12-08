import React from 'react'
import {doDelete, doPost} from "../common/http";

const notificationStyle = {
    position: "relative",
    width: "100%",
    padding: "6px",
    margin: "0px",
}
const notificationHeadingStyle = {
    fontSize: "1.2em",
    paddingBottom: "3px",
}

const notificationBodyStyle = {
    width: "100%",
}

const crossStyle = {
    position: "absolute",
    top: "10px",
    right: "20px",
    width: "15px",
    height: "15px",
}



export let Notification = (props) => {

    let dismiss = (event) => {
        doDelete(props.notification);
        event.stopPropagation()
    }

    let notificationClicked = (event) => {
        doPost(props.notification, {action: "default"});
        event.stopPropagation()
    }

    let {notification} = props

    // dangerouslySetInnterHtml should be safe here - we rely on
    // RefudeNotificationsService to sanitize notification body
    return (
        <div style={notificationStyle} onClick={notificationClicked}>
            <div style={notificationHeadingStyle}>{notification.Subject}</div>
            <div style={notificationBodyStyle} dangerouslySetInnerHTML={{__html: notification.Body}}/>

            {Object.keys(notification._actions).filter(k => k !== "default").map(k => {
                let buttonClicked = (event) => {
                    doPost(props.notification, {action: k}, "POST")
                    event.stopPropagation()
                }

                return <input type="submit" value={notification._actions[k].Description} onClick={buttonClicked}/>
            })}

            <div style={crossStyle} onClick={dismiss}>
                <svg height="15px" width="15px" viewBox="0 0 100 100"
                     strokeLinecap="round" stroke="grey" strokeWidth="10">
                    <circle cx="50" cy="50" r="40" fill="none"/>
                    <line x1="32" y1="32" x2="68" y2="68"/>
                    <line x1="32" y1="68" x2="68" y2="32"/>
                </svg>
            </div>
            <hr/>
        </div>
    )
};

