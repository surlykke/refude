
chrome.app.runtime.onLaunched.addListener(function (launchData) {
    chrome.app.window.create('index.html', {
        'id': 'org.restfulipc.refude.DefaultApps',
        'outerBounds': {
            'width': 460,
            'height': 300,
            'minWidth': 460,
            'minHeight': 300
        }
    });
});