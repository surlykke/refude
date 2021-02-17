import React, { useState, useEffect } from "react"
import { path2Url, iconClassName } from "../common/monitor"


export let move = up => {
    let ae = document.activeElement
    let candidate = ae && up ? ae.previousElementSibling : ae.nextElementSibling
    if (!candidate || !candidate.classList.contains("item")) {
        let items = Array.from(document.getElementsByClassName("item"))
        candidate = items[up ? items.length - 1 : 0]
    }
    candidate && candidate.focus()
}

export let LinkList = ({ links, focusedHref, selectLink, act}) => {
    useEffect(() => {
        let element = (focusedHref && document.getElementById(focusedHref)) || (links[0] && document.getElementById(links[0].href))
        element && element.focus()
    })

    return <div className="itemlist">
        {links.map((l, i) =>
            <div key={l.href} id={l.href}
                className="item"
                onFocus={() => selectLink(l)}
                onDoubleClick={act}
                tabIndex={i + 1} >
                {l.icon && <img className={iconClassName(l)} src={path2Url(l.icon)} height="20" width="20" />}
                <div className="title"> {l.title}</div>
            </div>)}
    </div>
}