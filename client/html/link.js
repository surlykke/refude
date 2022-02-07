// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import {doDelete, doPost, iconClassName} from "./utils.js"
import { img, a, span } from "./elements.js"

export let link = (link, comment, dismiss, move) => {
    
    let  onKeyDown = event => {
        console.log("link, onKeyDown", event)
        let { key, ctrlKey, shiftKey, altKey} = event;
        if (key === "ArrowRight" || key === "l" && ctrlKey) {
            if (event.target.rel === "related") {
                move("right", event.target.href);
            }
        } else if (key === "ArrowLeft" || key === "h" && ctrlKey) {
            move("left")
        } else if (key === "ArrowUp" || key === "k" && ctrlKey || key === 'Tab' && shiftKey && !ctrlKey && !altKey) {
            move("up");
        } else if (key === "ArrowDown" || key === "j" && ctrlKey || key === 'Tab' && !shiftKey && !ctrlKey && !altKey) {
            move("down");
        } else if (key === "Enter") {
            doPost(event.target.href).then(response => response.ok && !ctrlKey && dismiss())
        } else if (key === "Delete") {
            doDelete(event.target.href).then(response => response.ok && !ctrlKey && dismiss())
        } else { 
            return;
        }
        event.preventDefault();
    }

    comment = comment || "" 
    return a({  className: "link", 
                onClick: e => e.preventDefault(),
                onDoubleClick: e => doPost(e.target.href).then(response => response.ok && dismiss && dismiss()),
                onKeyDown: onKeyDown,
                rel:link.rel, 
                href: link.href,
                tabIndex: -1,
             }, 
        link.icon && img({className: iconClassName(link.profile), src:link.icon, height:"20", width:"20"}), 
        span({className:"title"}, link.title),
        span({className:"comment"}, comment)
    )
}

