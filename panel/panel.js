/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

var setclock = function() {
    var now = new Date();
    document.getElementById("time").innerHTML = now.toLocaleTimeString();
    setTimeout(setclock, 1000 - now.getMilliseconds() + 3); // Just after next turn of second..
};

document.addEventListener('DOMContentLoaded', setclock);
