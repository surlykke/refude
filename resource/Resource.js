import React, { useEffect } from "react"
import { path2Url, iconClassName } from "../common/monitor"
import { GenericResource } from "./GenericResource"


export let move = up => {
    let ae = document.activeElement
    let candidate = ae && up ? ae.previousElementSibling : ae.nextElementSibling
    if (!candidate || !candidate.classList.contains("item")) {
        let items = Array.from(document.getElementsByClassName("item"))
        candidate = items[up ? items.length - 1 : 0]
    }
    candidate && candidate.focus()
}


let component = ({ _self }) => {
    console.log("_self:", _self)
    if (_self.traits && _self.traits.includes("search")) {
        return null
    } else {
        return <GenericResource self={_self} />
    }

}

let search = (resource, term) => !resource.self && term &&
    <div className="searchbox" >
        <i className="material-icons" style={{ color: "lightGrey" }}>search</i>
        {term}
    </div>

export let Resource = ({ resource, controller, term }) => {
    if (!resource) {
        return null
    } else {

        let actions = resource._self.options.POST || []
        let links = resource._related
        useEffect(() => controller.focus(links))
        let tabindex = 1
        return <>
            {component(resource)}
            {search(resource, term)}
            <div className="itemlist">
                {actions.map(a =>
                    <div key={"act:" + a.actionId}
                        className="item"
                        onFocus={() => controller.onFocus(a, resource._self)}
                        onDoubleClick={() => controller.activate(actionId)}
                        tabIndex={tabindex++} >
                        {a.icon && <img className="icon" src={path2Url(a.icon)} height="20" width="20" />}
                        <div className="title"> {a.title}</div>
                    </div>
                )}
                {links.map((l, i) =>
                    <div key={l.href} id={l.href}
                        className="item"
                        onFocus={() => controller.onFocus(l)}
                        onDoubleClick={controller.activate}
                        tabIndex={tabindex++} >
                        {l.icon && <img className={iconClassName(l)} src={path2Url(l.icon)} height="20" width="20" />}
                        <div className="title"> {l.title}</div>
                    </div>)}
            </div>
        </>
    }
}