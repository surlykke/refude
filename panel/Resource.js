import React, { useEffect } from "react"
import { path2Url, iconClassName} from "../common/monitor"

export let Resource = ({resource}) => {
    if (resource.traits && resource.traits.includes("search")) {
        return null
    } else {
        let self = resource._links.find(l => "self" === l.rel)
        return <div key="resource" id={self.href} className="self">
            <img width="32px" height="32px" className={iconClassName(self)} src={path2Url(self.icon)} alt="" />
            <div className="name">{self.title}</div>
        </div> 
    }

}

export let Term = ({term}) => 
    term ?
    <div className="searchbox" >
        <i className="material-icons" style={{ color: "lightGrey" }}>search</i>
        {term}
    </div> : null

let setFocus = () => {
    let items = document.getElementsByClassName("item")
    if (items.length >0) {
        console.log("focus", items[0])
        items[0].focus()
    } else {
        console.log("focus",document.getElementById("itemList"))
        document.getElementById("itemList").focus()
    }
}


export let Links = ({links, activate, select}) => {
    useEffect(setFocus)
    return <div className="itemlist" id="itemList">
        {links.map((l, i) =>
            <div key={l.href} id={l.href} data-url-={l.href}
                className="item"
                onDoubleClick={() => activate(l, true)}
                onFocus={() => select(l)} 
                tabIndex={i + 1} >
                {l.icon && <img className={iconClassName(l)} src={path2Url(l.icon)} height="20" width="20" />}
                <div className="title"> {l.title}</div>
            </div>)
        }
    </div>
}