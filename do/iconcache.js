/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

makeIconCache = function($http, resource) {
    var iconRequestQueue = [];
    fetcherIsWorking = false;

    var cache = {
        urls: {},
        requestIcon: function (iconName) {
            iconRequestQueue.push(iconName);
            if (!fetcherIsWorking) {
                fetcher();
            }
        }
    };

    function fetcher() {
        if (iconRequestQueue.length > 0) {
            fetcherIsWorking = true;
            var iconName = iconRequestQueue.shift();
            if (iconName in cache.urls) {
                // Move on to next request, then 
                fetcher();
            } 
            else {
                var url = resource + "?name=" + iconName;
                console.log("fetch icon from url: ", url); 
                $http.get(url, {responseType: 'blob'}).then(
                        function success(response) {
                            cache.urls[iconName] = window.URL.createObjectURL(response.data);
                            fetcher();
                        },
                        function error(response) {
                            cache.urls[iconName] = null;
                            fetcher();
                        }
                );
            }
        }
        else {
            fetcherIsWorking = false;
        }
    };

    return cache;
};
    