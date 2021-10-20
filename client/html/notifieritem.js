// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { div, doPost, img, linkHref, menuHref } from './utils.js'
import {menu} from './menu.js'

let NotifierItem = ({res}) => {


    const [menuObject, setMenuObject] = React.useState(undefined)

    let href = linkHref(res)
    let mHref = menuHref(res)

    let getXY = (event) => {
        return {
            x: Math.round(event.view.devicePixelRatio * event.screenX),
            y: Math.round(event.view.devicePixelRatio * event.screenY)
        }
    }

    let click = (event) => {
        event.persist()
        event.preventDefault()
        console.log("notifierItem click:", event.target, event.currentTarget)
        let { x, y } = getXY(event)
        if (event.button === 0) {
            fetch(`${href}?action=left&x=${x}&y=${y}`, {method: "POST"});
        } else if (event.button === 1) {
            fetch(`${href}?action=middle&x=${x}&y=${y}`, {method: "POST"});
        }
    }


    let rightClick = (event) => {
        event.persist()
        event.preventDefault()
        if (mHref) {
            fetch(mHref) 
                .then(resp => resp.json())
                .then(json => setMenuObject(json),
                      ()   => setMenuObject(undefined))
        } else {
            let { x, y } = getXY(event)
            doPost(`${href}?action=right&x=${x}&y=${y}`)
        }
    }

    let onKeyDown = (event) => {
        if (event.key === "Escape") {
            event.preventDefault();
            setMenuObject(undefined)
        } 
    }

    return (
        div(
            {
                className:"clickable", 
                onKeyDown: onKeyDown, 
            }, 
            img({src:res.icon, alt:"", height:"20px", width:"20px", onClick:click, onContextMenu:rightClick}),
            menuObject && menu(menuObject, () => setMenuObject(undefined))
        )
    )
}

export let notifierItem = res => React.createElement(NotifierItem, {res: res})

