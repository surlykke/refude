document.addEventListener("dismiss", () => {
    chrome.runtime.sendMessage("dismiss")
});