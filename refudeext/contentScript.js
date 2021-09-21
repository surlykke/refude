document.addEventListener("dismiss", () => {
    chrome.runtime.sendMessage("dismiss")
});

chrome.runtime.onMessage.addListener(cmd => document.dispatchEvent(new Event(cmd)))