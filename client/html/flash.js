import { div, img } from "./elements.js";

export let flash = n => {
    console.log("flash:", n)
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
