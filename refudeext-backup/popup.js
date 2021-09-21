console.log("Loader popup.js")
const refudeUrl = 'http://localhost:7938/client/base.html'
let openButton = document.getElementById('open')
let closeButton = document.getElementById('close')

openButton.addEventListener("click", async () => {
    console.log("Hej fra open");
    findRefudeTabAndThen(tab => {
        console.log(tab)
        tab ? chrome.tabs.highlight({'tabs': tab.index})
            : chrome.tabs.create({url: refudeUrl})
    }) 
});

closeButton.addEventListener("click", async () => {
    findRefudeTabAndThen(tab => {
        if (tab) {
            console.log("tab:", tab)
            chrome.tabs.remove(tab.id)
        }
    })
});

let findRefudeTabAndThen = handler => {
    chrome.tabs.query({url: "http://localhost/*"}, tabs => {
        handler(tabs.find(t => t.url === refudeUrl))
    })
}
