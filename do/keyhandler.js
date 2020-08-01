let noop = () => { }

export let keyDownHandler = (up, down, left, right, post, del, dismiss) => {
    up = up || noop
    down = down || noop
    left = left || noop
    right = right || noop
    post = post || noop
    del = del || noop
    dismiss = dismiss || noop

    return event => {
        let { key, ctrlKey, shiftKey, altKey, metaKey } = event;

        if ((key === "Tab" && !ctrlKey && shiftKey && !altKey && !metaKey) ||
            (key === "ArrowUp" && !ctrlKey && !shiftKey && !altKey && !metaKey) ||
            (key === "k" && ctrlKey && !shiftKey && !altKey && !metaKey)) {
            up()
        } else if ((key === "Tab" && !ctrlKey && !shiftKey && !altKey && !metaKey) ||
            (key === "ArrowDown" && !ctrlKey && !shiftKey && !altKey && !metaKey) ||
            (key === "j" && ctrlKey && !shiftKey && !altKey && !metaKey)) {
            down()
        } else if ((key === "ArrowLeft" && !ctrlKey && !shiftKey && !altKey && !metaKey) ||
            (key === "h" && ctrlKey && !shiftKey && !altKey && !metaKey)) {
            left()
        } else if ((key === "ArrowRight" && !ctrlKey && !shiftKey && !altKey && !metaKey) ||
            (key === "l" && ctrlKey && !shiftKey && !altKey && !metaKey)) {
            right();
        } else if ((key === "Enter" && !ctrlKey && !shiftKey && !altKey && !metaKey) ||
            (key === " " && !ctrlKey && !shiftKey && !altKey && !metaKey)) {
            post();
        } else if (key === "Delete" && !ctrlKey && !shiftKey && !altKey && !metaKey) {
            del();
        } else if (key === "Escape" && !ctrlKey && !shiftKey && !altKey && !metaKey) {
            dismiss();
        } else {
            return;
        }
        event.preventDefault();
    }
}