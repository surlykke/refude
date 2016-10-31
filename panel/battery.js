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
        log("battery update");
        let style ="color: black;";
        if (["Charging", "Fully charged"].indexOf(state) > -1) { 
            element.innerHTML = "<b>" + charge + "%</b>";
        }
        else if (["Discharging", "Empty"].indexOf(state) > -1) {
            element.innerHTML = "" + charge + "%";
            if (charge < 20) {
                style = "color: red;";
            }
        }
        else {
            element.innerHTML = "?";
        }
        element.style = style;
    };

    let updateBatteryInfo = function (event) {
        log("updateBatteryInfo");
        GET({host: "localhost", port: 7938, path: "/power-service/devices/DisplayDevice"}, function (json) {
            log("updateBatteryInfo, return")
            update(json.State, json.Percentage);
        });
    };

    log("Creating evtSource");
    let evtSource = new EventSource("http://localhost:7938/power-service/notify");
   
    log("Set onerror");
    evtSource.onerror = function (event) {
        log("battery error")
        //console.log("evtSource error", event);
        update("Unknown", 0);
    };

    log("Set onopen");
    evtSource.onopen = function (event) {
        log("battery open")
        //console.log("evtSource onopen", event);
        updateBatteryInfo();
    };

    log("addEventListener");
    evtSource.addEventListener("resource-updated", function (e) {
        //console.log("evtSource update", e);
        updateBatteryInfo();
    });   
});


