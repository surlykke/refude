// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.

export const div = (props, ...children) => React.createElement('div', props, ...children)
export const p = (props, ...children) => React.createElement('p', props, ...children) 
export const img = (props, ...children) => React.createElement('img', props, ...children) 
export const table = (props, ...children) => React.createElement('table', props, ...children)
export const tbody = (props, ...children) => React.createElement('tbody', props, ...children)
export const tr = (props, ...children) => React.createElement('tr', props, ...children)
export const td = (props, ...children) => React.createElement('td', props, ...children)
export const materialIcon = (name) => React.createElement('i', {className: 'material-icons', style:{color: 'light-grey'}}, name)
export const frag = (...children) => React.createElement(React.Fragment, null, ...children)

export const iconClassName = profile => "icon" + ("window" === profile ? " window" : "")

export const getJson = href => fetch(href).then(resp => resp.json())
export const doPost = (href, paramMap) => {
    if (paramMap) {
        let separator = href.indexOf('?') > -1 ? '&' : '?'
        for (const [name, val] of Object.entries(paramMap)) {
            href = href + separator + name + "=" + val
            separator = "&"
        }
    }
    return fetch(href, {method: "POST"})
}

export const doDelete = href => fetch(href, {method: "DELETE"})

export let linkHref = (res, rel) => {
    rel = rel || "self"
    return res._links.find(l => l.rel === rel)?.href
}
export let menuHref = res => linkHref(res, "org.refude.menu") 

