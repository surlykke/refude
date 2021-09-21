import {div, img, iconClassName} from "./utils.js"

let Links = ({links, activate, select}) => {
    React.useEffect(setFocus) 
    let [html, actionJustPushed, tabIndex] = [[], false, 1]
    links.forEach((l, i) => {
        if (l.rel.endsWith('action')) {
            actionJustPushed || html.push(div({key:"actionH", className:'itemheading'}, "Actions"))
            html.push(link({key:l.href,link:l,tabIndex:tabIndex++, activate:activate, select:select}))
            actionJustPushed = true
        } else if (l.rel === "related") {
            actionJustPushed && html.push(div({ key:"relatedH", className:'itemheading'}, "Related"))
            html.push(link({key:l.href, link:l, tabIndex:tabIndex++, activate:activate, select:select}))
            actionJustPushed = false     
        }
    })
    
    return div({id:"itemList"}, html)
}

export let links = (resource, activate, select) => 
    React.createElement(Links, {links: resource._links, activate: activate, select: select})

let Link = ({link, tabIndex, activate, select}) => {
    let act = () => activate(link, true)
    let cn = iconClassName(link.profile)
    return ( 
        div(
            {id: link.href, className:"item", tabIndex:tabIndex, onDoubleClick: act, onFocus:() => select(link)}, 
            link.icon && img({className:cn, src:link.icon, height:"20", width:"20"}), 
            div({className:"title"}, link.title)
        )
    )
}

let link = props => React.createElement(Link, props) 

let setFocus = () => {
    let items = document.getElementsByClassName("item")
    if (items.length >0) {
        items[0].focus()
    } else {
        document.getElementById("itemList").focus()
    }
}

