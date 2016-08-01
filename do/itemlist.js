/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
var makeItemList = function ($http, iconCache) {
    var selectedIndex = -1;
    var obj = {
        windows: [],
        applications: [],
        filteredWindows: [],
        filteredItems: [],
        selectedUrl: function() { 
            return selectedIndex === -1 ? undefined : obj.filteredItems[selectedIndex].url;
        },
        next: function() { 
            if (selectedIndex > -1) selectedIndex = (selectedIndex + 1) % obj.filteredItems.length; 
        },
        previous: function() { 
            if (selectedIndex > -1) {
                selectedIndex = (selectedIndex + obj.filteredItems.length - 1) % obj.filteredItems.length;
            }
        },
        filter: function(searchTerm) {
            var selectedUrl = obj.selectedUrl();

            var searchTerm = searchTerm ? searchTerm.trim() : "";
            if (searchTerm === "") {
                obj.filteredItems = obj.windows;
                obj.filteredWindows = obj.windows;
            }
            else { 
                obj.filteredItems = [];
                obj.filteredWindows = [];
                obj.windows.forEach(function(window) {
                    if (window.name.toLowerCase().includes(searchTerm.toLowerCase())) {
                        obj.filteredItems.push(window);
                        obj.filteredWindows.push(window);
                    }
                });
                obj.applications.forEach(function(app) {
                    if (app.name.toLowerCase().includes(searchTerm.toLowerCase())) {
                        obj.filteredItems.push(app);
                    }
                });
            }
        
            selectedIndex = obj.filteredItems.findIndex(item=>item.url === selectedUrl);
            if (selectedIndex === -1 && obj.filteredItems.length > 0) {
                selectedIndex = 0;
            }
            obj.filteredItems.forEach(function(item) {
                iconCache.requestIcon(item.iconUrl);
            });
        } 
    };

    

    var getWindows = function() {
        var windowsUrl = "http://localhost:7938/wm-service/windows";
        $http.get(windowsUrl).then(function(response) {
            obj.windows = response.data.map(function(window) {
                return  {
                    name: window.Name,
                    id: window.Id,
                    comment: "running",
                    isAWindow: true,
                    geometry: window.geometry,
                    url: "http://localhost:7938/wm-service/windows/" + window.Id,
                    iconUrl: "http://localhost:7938/wm-service/icons/" + window.windowIcon
                };
            });
            obj.filter();
        });
    };
    
    var evtSource = new EventSource("http://localhost:7938/wm-service/notify");

    var eventHandler = function(event) {
        getWindows();
    };

    $http.get("http://localhost:7938/desktop-service/applications").then(function(response) {
        obj.applications = response.data.map(function(app) {
            return { 
                name: app.Name,
                id : app.applicationId,
                comment: app.Comment, 
                url : "http://localhost:7938/desktop-service/applications/" + app.applicationId,
                iconUrl: "http://localhost:7938/icon-service/icons/icon?name=" + app.Icon
            };
        });

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
    });

    return obj;
};