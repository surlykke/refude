import { doPost, linkHref } from "./utils.js"
import { div} from "./elements.js"

const ul = (...children) => {
    return React.createElement('ul', undefined, ...children)
}

const toggleVisible = event => {
    event.currentTarget.classList.toggle('open')
}

let Menu = ({menuObject, dismiss}) => {
    let makeClickHandler = (id) => () => doPost(linkHref(menuObject), {id: id}).then(() => dismiss())

    let entry = e => {
        let { Label: text, Type, ToggleState, Id, SubEntries } = e
        text = (text || "").replace(/_([^_])/g, "$1")
        let clickHandler

        if (Type === 'separator') {
            return React.createElement('hr')
        } else {
            let className = ""
            if (SubEntries) {
                className = "submenu"
                clickHandler = toggleVisible
            } else {
                if (ToggleState > 0) {
                    className = "marked"
                }
                clickHandler = makeClickHandler(Id)
            } 
            return React.createElement('li', 
                {className: className, onClick: clickHandler},
                React.createElement('span', {}, text),
                SubEntries && entryList(SubEntries) 
            )
        }
    }

    let entryList = (entries) => {
        return ul(...(entries.map(e => entry(e))))
    }

    return div({className: 'menu'}, entryList(menuObject.data))
}

export const menu = (menuObject, dismiss) => React.createElement(Menu, { menuObject: menuObject, dismiss: dismiss })