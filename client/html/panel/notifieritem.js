// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { doPost, linkHref, menuHref } from '../common/utils.js'
import { div, img } from "../common/elements.js"
import {menu} from './menu.js'

let NotifierItem = ({res, setMenuObject}) => {

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
                onKeyDown: onKeyDown, 
            }, 
            img({src:res.icon, alt:"", height:"16px", width:"16px", onClick:click, onContextMenu:rightClick})
        )
    )
}

export let notifierItem = (res, setMenuObject) => 
    React.createElement(NotifierItem, {res: res, setMenuObject: setMenuObject})

