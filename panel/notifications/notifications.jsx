// Copyright (c) 2015, 2016, 2017 Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import {render} from 'react-dom'
import {doGet, doPost, doDelete} from '../../common/utils'

const notificationStyle = {
    position: "relative",
    maxWidth: "200px",
    padding: "6px",
    margin: "0px",
    backgroundColor: "lightgrey",
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
    top: "3px",
    right: "2px",
    width: "15px",
    height: "15px",
}

let Notification = (props) => {

    let dismiss = (event) => {
        console.log(event)
        doDelete(props.item);
        event.stopPropagation()
    }

    let notificationClicked = (event) => {
        console.log("notification clicked");
        doPost(props.item.Self, {action: "default"});
        event.stopPropagation()
    }

    let {item} = props

    // dangerouslySetInnterHtml should be safe here - we rely on
    // RefudeNotificationsService to sanitize notification body
    return (
        <div style={notificationStyle} onClick={notificationClicked}>
            <div style={notificationHeadingStyle}>{item.Subject}</div>
            <div style={notificationBodyStyle} dangerouslySetInnerHTML={{__html: item.Body}}/>

            {Object.keys(item.Actions).filter(k => k !== "default").map(k => {
                let buttonClicked = (event) => {
                    console.log("button", k, "clicked")
                    doPost(props.item, {action: k}, "POST")
                    event.stopPropagation()
                }

                return <input type="submit" value={item.Actions[k]} onClick={buttonClicked}/>
            })}

            <div style={crossStyle} onClick={dismiss}>
                <svg height="15px" width="15px" viewBox="0 0 100 100"
                     strokeLinecap="round" stroke="grey" strokeWidth="10">
                    <circle cx="50" cy="50" r="40" fill="none"/>
                    <line x1="32" y1="32" x2="68" y2="68"/>
                    <line x1="32" y1="68" x2="68" y2="32"/>
                </svg>
            </div>

        </div>
    )
}


class Notifications extends React.Component {
    constructor(props) {
        super(props)
        this.state = {items: []}
        this.onUpdated = props.onUpdated
        console.log("constructor: this.state.items:", this.state.items)
    }

    componentDidMount = () => {
        let update = () => {
            doGet("notifications-service", "/search", {type: "application/vnd.org.refude.desktopnotification+json"}).then(notifications => {
                this.setState({items: notifications});
            }).catch().then(setTimeout(update, 1000));
        };
        update();
    }

    componentDidUpdate = () => {
        this.onUpdated()
    }

    render = () => {
        console.log("render Notifications, items: ", this.state.items)
        return (<div>
            {this.state.items.map(item => (<Notification key={item.Id} item={item}/>))}
        </div>);
    }
}

export {Notifications}
