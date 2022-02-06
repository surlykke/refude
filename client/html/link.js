// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import {iconClassName} from "./utils.js"
import { img, a, span } from "./elements.js"
import { selectLink } from "./navigation.js"

const click = e => {
    e.preventDefault()
    selectLink(e.currentTarget)
}

export let link = (link, comment, dblClick) => {
    comment = comment || "" 
    return a({className: "link", onClick: click, onDoubleClick: dblClick, rel:link.rel, href: link.href}, 
        link.icon && img({className: iconClassName(link.profile), src:link.icon, height:"20", width:"20"}), 
        span({className:"title"}, link.title),
        span({className:"comment"}, comment)
    )
}

