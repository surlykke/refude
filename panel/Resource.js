import React, { useEffect } from "react"

let iconClassName = profile => "icon" + ("window" === profile ? " window" : "")

export let Resource = ({resource, term, activate, select}) => 
    <>
        <ResourceHead resource={resource}/>
        <Term term={term}/>
        <Links links={resource._links} activate={activate} select={select}/>
    </>

export let ResourceHead = ({resource}) => {
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

    return <> 
        <div className="resource-header">
            <div className="resource-icon">
                {icon && <img className={iconClassName(profile)} src={icon} height="32" width="32" />}
            </div>
            <div>
                <div className="resource-name"> {title}</div>
                <div className="resource-comment"> {comment}</div>
            </div>
        </div>
        
        { props && 
        <table className="resource-props">
            <tbody>
                {Object.keys(props).map(key => 
                <tr className="prop">
                    <td className="key">{key}:</td>
                    <td className="value">{props[key]}</td>
                </tr>
                )}
            </tbody>
        </table>}
    </>
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
        items[0].focus()
    } else {
        document.getElementById("itemList").focus()
    }
}

export let Links = ({links, activate, select}) => {
    useEffect(setFocus) 
    let [html, actionJustPushed, tabIndex] = [[], false, 1]
    links.forEach((l, i) => {
        if (l.rel.endsWith('action')) {
            actionJustPushed || html.push(<div key="actionH" className='itemheading'>Actions</div>)
            html.push(<Link key={l.href}  link={l} tabIndex={tabIndex++} activate={activate} select={select}/>)
            actionJustPushed = true
        } else if (l.rel === "related") {
            actionJustPushed && html.push(<div key="relatedH" className='itemheading'>Related</div>)
            html.push(<Link key={l.href}  link={l} tabIndex={tabIndex++} activate={activate} select={select}/>)
            actionJustPushed = false     
        }
    })
    
    return <div className="itemlist" id="itemList">
        {html}
    </div>
}

export let Link = ({link, tabIndex, activate, select}) =>  
    <div id={link.href} data-url-={link.href}
        className="item"
        tabIndex={tabIndex}
        onDoubleClick={() => activate(link, true)}
        onFocus={() => select(link)}>
        {link.icon && <img className={iconClassName(link.profile)} src={link.icon} height="20" width="20" />}
        <div className="title"> {link.title}</div>
    </div>