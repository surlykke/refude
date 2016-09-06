/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
let  state = 4;
let charge = 0;
let stateStr = "";
let batteryclass = "battery-good";
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

let update = function () {
    let stateChar = ["\u25CF", "\u002B", "\u2212", "\u25CB", "?"][state];
    stateStr = "" + charge + "% (" + stateChar + ")";

    if (state < 2 || charge > 20) {
        batteryclass = "battery-good";
    } else if (charge > 10) {
        batteryclass = "battery-low";
    } else {
        batteryclass = "battery-critical";
    }
    document.getElementById("state").innerHTML = stateStr;
    document.getElementById("battery").className = batteryclass;
    document.getElementById("battery").style.width = "" + charge + "%";
};
var updateBatteryInfo = function (event) {
    GET({host: "localhost", port: 7938, path: "/power-service/devices/DisplayDevice"}, function (json) {
        console.log("Got json:", json)
        charge = json.Percentage;
        state = json.State;
        console.log("charge: ", charge, ", state: ", state);
        update();
    });
};
var evtSource = new EventSource("http://localhost:7938/power-service/notify");
evtSource.onerror = function (event) {
    console.log("Error:", event);
    charge = 0;
    state = 4;
    update();
};
evtSource.onopen = function (event) {
    console.log("open", event);
    updateBatteryInfo();
};
evtSource.addEventListener("resource-updated", function (e) {
    updateBatteryInfo();
});
let openPowerSettings = function () {
    let desktopClient = requestJson.createClient("http://localhost:7938/desktop-service/applications/lxqt-leave.desktop");
    desktopClient.post();
};
