// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import React from 'react'
import Axios from 'axios'
import { remote } from 'electron'
const { Menu, MenuItem } = remote

export let NotifierItem = ({ itemLink }) => {

    let showMenu = (menuHref) => {
        let clickHandler = (id) => {
            return () => { Axios.post(`${menuHref}?id=${id}`) }
        }

        let buildMenu = entries => {
            let menu = new Menu()
            entries.forEach(jsonMenuItem => {
                let label = (jsonMenuItem.Label || "").replace(/_([^_])/g, "$1")

                if (jsonMenuItem.SubEntries) {
                    menu.append(new MenuItem({ label: label, type: "submenu", submenu: buildMenu(jsonMenuItem.SubEntries) }))
                } else if (jsonMenuItem.Type === "separator") {
                    menu.append(new MenuItem({ type: "separator" }))
                } else if (jsonMenuItem.ToggleType === "checkmark") {
                    menu.append(new MenuItem({ label: label, type: "checkbox", click: clickHandler(jsonMenuItem.Id) }))
                } else if (jsonMenuItem.ToggleType === "radio") {
                    menu.append(new MenuItem({ label: label, type: "radio", click: clickHandler(jsonMenuItem.Id) }))
                } else {
                    menu.append(new MenuItem({ label: label, type: "normal", click: clickHandler(jsonMenuItem.Id) }))
                }

            })

            return menu
        }


        Axios.get(menuHref)
            .then(({data}) => buildMenu(data.Entries).popup())
    }

    let getXY = (event) => {
        return {
            x: Math.round(event.view.devicePixelRatio * event.screenX),
            y: Math.round(event.view.devicePixelRatio * event.screenY)
        }
    }

    let onClick = (event) => {
        event.persist()
        event.preventDefault()

        let { x, y } = getXY(event)
        if (event.button === 0) {
            Axios.post(`${itemLink.href}?action=left&x=${x}&y=${y}`);
        } else if (event.button === 1) {
            Axios.post(`${itemLink.href}?action=middle&x=${x}&y=${y}`);
        }
    }

    let onRightClick = (event) => {
        event.persist()
        event.preventDefault()
        Axios.get(itemLink.href) 
            .then(({data}) => {
                let menuLink = data._links.find(l => l.rel === 'org.refude.menu')
                if (menuLink) {
                    showMenu(menuLink.href)
                } else {
                    let selfLink = data._links.find(l => l.rel === "self")
                    let { x, y } = getXY(event)
                    Axios.post(selfLink.href + '?action=right&x=' + x + '&y=' + y)
                }
        })

    }

    return <div className="clickable">
        <img src={itemLink.icon} alt="" height="20px" width="20px" onClick={onClick} onContextMenu={onRightClick}/>
    </div>
}

//</img> /*onContextMenu={onRightClick} />

