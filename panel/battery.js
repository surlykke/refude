/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
const http = require('http');

let GET = function(opts, handler) {
    console.log("Doing get on ", opts, "with", http);
    http.get(opts, function(response) {
        let body = '';
        response.on('data', function(data) {
            console.log("Got data", data);
            body += data;
        })
        .on('error', function(err) {
            console.log("response error:", err);
        })
        .on('end', function() {
            json = JSON.parse(body);
            handler(json);
        });
    })
    .on('error', function(error) {
        console.log("request error: ", error);
    });
};

document.addEventListener("DOMContentLoaded", function() {
    let element = document.getElementById("battery");

    let update = function (state, charge) {
        element.innerHTML = "" + charge + "% (" + ["\u25CF", "\u002B", "\u2212", "\u25CB", "?"][state] +  ")";
    };

    let updateBatteryInfo = function (event) {
        GET({host: "localhost", port: 7938, path: "/power-service/devices/DisplayDevice"}, function (json) {
            update(json.State, json.Percentage);
        });
    };

    let evtSource = new EventSource("http://localhost:7938/power-service/notify");
    
    evtSource.onerror = function (event) {
        update(4, 0);
    };

    evtSource.onopen = function (event) {
        updateBatteryInfo();
    };

    evtSource.addEventListener("resource-updated", function (e) {
        updateBatteryInfo();
    });   
});


