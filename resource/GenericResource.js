import React from "react"
import { path2Url, iconClassName } from "../common/monitor"

export let GenericResource = ({ self }) => {
    console.log("GenericResource with", self)
    return <div key="resource" id={self.href} className="self">
                <img width="32px" height="32px" className={iconClassName(self)} src={path2Url(self.icon)} alt="" />
                <div className="name">{self.title}</div>
            </div> 

}


