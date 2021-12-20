import {iconClassName} from "./utils.js"
import { activateSelected, select, selectActivateAndDismiss } from "./navigation.js"
import { div, img, a } from "./elements.js"

let LinkDivs = ({links, dismiss}) => {
    let dblClick = e => {
        select(e.currentTarget)
        activateSelected(dismiss)
    }



    React.useEffect(select) 
    let [html, actionJustPushed, tabIndex] = [[], false, 1]
    links.forEach((link, i) => {
        if (link.rel.endsWith('action')) {
            actionJustPushed || html.push(div({key:"actionH", className:'itemheading'}, "Actions"))
            html.push(linkDiv({key:link.href,link:link, dblClick: dblClick}))
            actionJustPushed = true
        } else if (link.rel === "related") {
            actionJustPushed && html.push(div({ key:"relatedH", className:'itemheading'}, "Related"))
            html.push(linkDiv({key:link.href, link:link, dblClick: dblClick}))
            actionJustPushed = false     
        }
    })
    
    return div({className: "linkDivs"}, html)
}

export let linkDivs = (resource, dismiss) => React.createElement(LinkDivs, {links: resource._links, dismiss: dismiss})

let aOnClick = e => e.preventDefault()

let click = e => select(e.currentTarget)

let LinkDiv = ({link, dblClick}) => {
    return ( 
        div(
            {className: "link", onClick: click, onDoubleClick: dblClick},
            div({}, link.icon && img({src:link.icon, className: iconClassName(link.profile), height:"20", width:"20"})),
            a(
                {href: link.href, rel: link.rel, tabIndex: -1, onClick: aOnClick }, 
                link.title
            )
        )
    )
}

let linkDiv = props => React.createElement(LinkDiv, props) 


