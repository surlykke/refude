// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
import { div, img } from "./elements.js";

export let flash = n => {
    let size = 48
    if (n?.data.IconSize > 48) {
        size = Math.min(n.data.IconSize, 256)
    }
    return (
        div({ className: "flash" },
            div({ className: "flash-icon" },
                img({ height: `${size}px`, src: n.icon, alt: "" })
            ),
            div({ className: "flash-message" },
                div({ className: "flash-title" }, n.title),
                div({ className: "flash-body" }, n.comment)
            )
        )
    )
}
