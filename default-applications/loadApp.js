/*
* Copyright (c) 2015, 2016 Christian Surlykke
*
* This file is part of the refude project. 
* It is distributed under the GPL v2 license.
* Please refer to the LICENSE file for a copy of the license.
*/


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