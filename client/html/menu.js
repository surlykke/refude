import { doPost, linkHref } from "./utils.js"
import { div, p } from "./elements.js"

const ul = (props, ...children) => React.createElement('ul', props, ...children)
const menuItem = (text, marked, clickHandler) => {
    return React.createElement('li', 
        {},
        React.createElement('div', {className: 'menu-item'}, 
            React.createElement('span', {className: "marker"}, marked && "\u2713"),
            React.createElement('span', {className: 'label', onClick: clickHandler}, text))
    )
}

const subMenuItem = (text, subEntries) => {
    return React.createElement('li', 
        {},
        React.createElement('div', {className: 'menu-item'}, 
            React.createElement('span', {className: "marker"}),
            React.createElement('span', {className: 'label'}, text),
            React.createElement('span', {className: 'submarker'}, "\u25B8")),
        subEntries
    )
}

let Menu = ({menuObject, dismiss}) => {

    React.useEffect(() => document.getElementById('menu').focus())

    let makeClickHandler = (id) => () => doPost(linkHref(menuObject), {id: id}).then(() => dismiss())

    let entry = e => {
        let { Label: text, Type, ToggleState, Id, SubEntries } = e
        text = (text || "").replace(/_([^_])/g, "$1")

        if (SubEntries) {
            return subMenuItem(text, entryList(SubEntries))
        } else if (Type === "separator") {
            return React.createElement('hr')
        } else {
            return menuItem(text, ToggleState > 0, makeClickHandler(Id))
        }
    }

    let entryList = (entries) => {
        return ul({}, ...entries.map(e => entry(e)))
    }

    return div({className: 'menu', id:'menu', tabIndex: -1, onBlur: dismiss}, entryList(menuObject.data))
}

export const menu = (menuObject, dismiss) => React.createElement(Menu, { menuObject: menuObject, dismiss: dismiss })
