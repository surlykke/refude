import { div, img, table, tbody, tr, td } from "../common/elements.js"

let iconClassName = profile => "icon" + ("window" === profile ? " window" : "")

let ResourceHead = ({resource}) => {
    let {title, comment, icon, profile, data} = resource
    let props
    switch (profile) {
    case "search":
        return null
    case "file": 
        [title, comment] = [data.Path, data.Dir ? "Directory" : "File"]
        break
    case "device":
        [title, comment] = [data.Type, data.Model]
        props = {
            "State": data.State,
            "Energy, design": data.EnergyFullDesign,
            "Energy, full": data.EnergyFull,
            "Energy, now": data.Energy,
        }
        break
    } 

    return [ 
        div({className:"resource-header"},
            div({className:"resource-icon"},
                icon ? img({className:iconClassName(profile), src:icon, height:"32", width:"32"}) : null
            ),
            div({},
                div({className:"resource-name"}, title),
                div({className:"resource-comment"}, comment)
            )
        ),
        props ?
        table({className:"resource-props"}, 
            tbody({}, ...Object.keys(props).map(key => 
                tr({className:"prop"},
                    td({className:"key"},key + ':'),
                    td({className:"value"},props[key])
                )
            ))
        ): null
    ]
}

export let resourceHead = resource => React.createElement(ResourceHead, {resource: resource})




