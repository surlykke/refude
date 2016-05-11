/*
* Copyright (c) 2015, 2016 Christian Surlykke
*
* This file is part of the refude project. 
* It is distributed under the GPL v2 license.
* Please refer to the LICENSE file for a copy of the license.
*/

chrome.app.runtime.onLaunched.addListener(function (launchData) {
    chrome.app.window.create('do.html', {
        'id': 'org.restfulipc.refude.Do',
        'outerBounds': {
            'width': 300,
            'height': 300,
            'minWidth': 180,
            'minHeight': 200
        }
    });
});