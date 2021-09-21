export const  div = (props, ...children) => React.createElement('div', props, ...children)
export const img = (props, ...children) => React.createElement('img', props, ...children) 
export const table = (props, ...children) => React.createElement('table', props, ...children)
export const tbody = (props, ...children) => React.createElement('tbody', props, ...children)
export const tr = (props, ...children) => React.createElement('tr', props, ...children)
export const td= (props, ...children) => React.createElement('td', props, ...children)
export const materialIcon = (name) => React.createElement('i', {className: 'material-icons', style:{color: 'light-grey'}}, name)
export const frag = (...children) => React.createElement(React.Fragment, null, ...children)

export const iconClassName = profile => "icon" + ("window" === profile ? " window" : "")

export const getJson = href => fetch(href).then(resp => resp.json())
export const doPost = href => fetch(href, {method: "POST"})
export const doDelete = href => fetch(href, {method: "DELETE"})
