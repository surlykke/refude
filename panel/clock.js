/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
let startUpdatingTime = function(element) {
    let updateTime = function() {
        let now = new Date();
        element.innerHTML = now.toLocaleTimeString();
        setTimeout(updateTime, 1000 - now.getMilliseconds() + 1); // Just after next turn of second..
    };
    updateTime();
}

document.addEventListener("DOMContentLoaded", function() {
    console.log("log", log);
    log("startUpdatingTime...");
    startUpdatingTime(document.getElementById("clock"));
});
