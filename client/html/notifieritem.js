// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { doPost, linkHref} from './utils.js'
import { div, img } from "./elements.js"

let NotifierItem = ({res, setMenuObject}) => {

    let getXY = (event) => {
        return {
            x: Math.round(event.view.devicePixelRatio * event.screenX),
            y: Math.round(event.view.devicePixelRatio * event.screenY)
        }
    }

    let click = (event) => {
        event.persist()
        event.preventDefault()
        let { x, y } = getXY(event)
        if (event.button === 0) {
            fetch(`${res.self}?action=left&x=${x}&y=${y}`, {method: "POST"});
        } else if (event.button === 1) {
            fetch(`${res.self}?action=middle&x=${x}&y=${y}`, {method: "POST"});
        }
    }


    let rightClick = (event) => {
        event.persist()
        event.preventDefault()
        let menuRetreieved 
        fetch(res.links)
            .then(resp => resp.json())
            .then(linksJson => fetch(linkHref(linksJson, 'org.refude.menu')))
            .then(resp => resp.json()) 
            .then(menuJson => { 
                setMenuObject(menuJson)
                menuRetreieved = true
            },() => setMenuObject(undefined))
            .finally(() => {
                if (!menuRetreieved) {
                    let { x, y } = getXY(event)
                    doPost(`${res.self}?action=right&x=${x}&y=${y}`)
                }
            })
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

