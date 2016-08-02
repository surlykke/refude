/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
var makeWindowList = function ($http, somethingChangedCallback) {
    var obj = {
        filter: function(searchTerm) {
            searchTerm = (searchTerm ? searchTerm.trim() : "").toLowerCase();            
            var tmp = windows.filter(window => 
                window.name !== "Refude Do" &&
                !window.state.includes("Above") &&
                (searchTerm === "" || window.name.toLowerCase().includes(searchTerm)));
            return tmp;
       }
    };

    var  windows = [];

    var getWindows = function() {
        var windowsUrl = "http://localhost:7938/wm-service/windows";
        $http.get(windowsUrl).then(function(response) {
            windows = response.data.map(function(window) {
                return  {
                    name: window.Name,
                    id: window.Id,
                    state : window.State,
                    comment: "",
                    isAWindow: true,
                    geometry: window.geometry,
                    url: "http://localhost:7938/wm-service/windows/" + window.Id,
                    iconUrl: "http://localhost:7938/wm-service/icons/" + window.windowIcon
                };
            });
            somethingChangedCallback();
        });
    };

    var evtSource = new EventSource("http://localhost:7938/wm-service/notify");

    var eventHandler = function(event) {
        getWindows();
    };

    evtSource.onerror = function(event) {
        obj.windows = [];
        obj.filter();
    };

    evtSource.onopen = function(event) {
        getWindows();
    }; 

    evtSource.addEventListener("resource-updated", eventHandler);
    evtSource.addEventListener("resource-added", eventHandler);
    evtSource.addEventListener("resource-removed", eventHandler);

    return obj;
};
