import { div, img } from "./elements.js";

export let flash = n => {
    return (
        div({ className: "flash" },
            div({ className: "flash-icon" },
                img({ width: "100%", height: "100%", src: n.icon, alt: "" })
            ),
            div({ className: "flash-message" },
                div({ className: "flash-title" }, n.title),
                div({ className: "flash-body" }, n.comment)
            )
        )
    )
}
