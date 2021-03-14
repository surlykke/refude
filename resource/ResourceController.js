import { deleteUrl, postUrl, addParam } from "../common/monitor"

export let makeResourceController = (navigateTo, navigateBack, dismiss, handleKey) => {
    let currentAction
    let currentLink

    let move = up => {
        let ae = document.activeElement
        let candidate = ae && up ? ae.previousElementSibling : ae.nextElementSibling
        if (!candidate || !candidate.classList.contains("item")) {
            let items = Array.from(document.getElementsByClassName("item"))
            candidate = items[up ? items.length - 1 : 0]
        }
        candidate && candidate.focus()
    }



    let onFocus = (item, resourceSelf) => {
        console.log("onFocus, item:", item, ",resourceSelf:", resourceSelf)
        if (item?.actionId) {
            currentAction = item
            currentLink = resourceSelf
        } else {
            currentAction = undefined
            currentLink = item
        }
        if (currentLink?.traits && currentLink.traits.includes("window")) {
            postUrl(addParam(currentLink.href, "actionId", "highlight"))
        } else {
            postUrl("/window/unhighlight")
        }
    }

    let focus = (links) => {
        let href = currentLink?.href
        let element = (href && document.getElementById(href)) || (links[0] && document.getElementById(links[0].href))
        element && element.focus()
    }

    let activate = then => {
        console.log("activate, then:", then, ",currentLink:", currentLink, "currentAction:", currentAction)
        if (!currentLink) {
            return
        }
        let href = currentLink.href
        if (currentAction) {
            href = addParam(href, "actionId", currentAction.actionId)
        }
        console.log("will post:", href)
        postUrl(href, () => then && then())
    }

    let activateAndDismiss = () => {
        postUrl("/window/unhighlight", () => activate(dismiss))
    }

    let doDelete = keep => {
        if (currentLink && !currentAction) {
            deleteUrl(currentLink.href, () => keep || dismiss())
        }
    }

    let onKeyDown = (event) => {
        let { key, keyCode, ctrlKey, altKey, shiftKey, metaKey } = event;
        if (key === "i" && ctrlKey && shiftKey) {
            ipcRenderer.send("devtools")
        } else if (key === "ArrowRight" || key === "l" && ctrlKey && !currentAction) {
            navigateTo(currentLink)
        } else if (key === "ArrowLeft" || key === "h" && ctrlKey) {
            navigateBack()
        } else if (key === "ArrowUp" || key === "k" && ctrlKey) {
            move(true)
        } else if (key === "ArrowDown" || key === "j" && ctrlKey) {
            move(false)
        } else if (key === "Enter") {
            ctrlKey ? activate() : activateAndDismiss()
        } else if (key === "Delete") {
            doDelete(ctrlKey)
        } else if (key === "Escape") {
            postUrl("/window/unhighlight", dismiss)
        } else if ((keyCode === 8 || key.length === 1) && !ctrlKey && !altKey && !metaKey) {
            handleKey(key)
        } else {
            return
        }
        event.preventDefault();
    }



    return {
        focus: focus,
        onFocus: onFocus,
        focused: () => currentlyFocused,
        onKeyDown: onKeyDown,
        activate: activate
    }

}
