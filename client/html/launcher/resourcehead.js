// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { div, img, table, tbody, tr, td } from "../common/elements.js"

let heading = (title, comment, icon, iconClassName) =>
    div({ className: "resource-header" },
        div({ className: "resource-icon" },
            icon ? img({ className: iconClassName, src: icon, height: "32", width: "32" }) : null
        ),
        div({},
            div({ className: "resource-name" }, title),
            comment && div({ className: "resource-comment" }, comment)
        )
    )


let device = resource => [
    heading(resource.data.Type, resource.data.Model, resource.icon),
    table({className:"resource-props"}, 
        tbody({},
            tr({}, td({}, "State"),          td({}, resource.data.State)),
            tr({}, td({}, "Energy, design"), td({}, resource.data.EnergyFullDesign)),
            tr({}, td({}, "Energry, full"),  td({}, resource.data.EnergyFull)),
            tr({}, td({}, "Energy, now"),    td({}, resource.data.Energy)),
            tr({}, td({}, "Percentage"),     td({}, resource.data.Percentage)),
        )
    )
]


export let resourceHead = resource =>  
    resource.profile === "start"  ? [] :
    resource.profile === "device" ? device(resource) : 
    resource.profile === "window" ? heading(resource.title, resource.comment, resource.icon, "window") :
                                    heading(resource.title, resource.comment, resource.icon)





