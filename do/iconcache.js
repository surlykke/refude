/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

makeIconCache = function ($http, resource) {

    var cache = {
        urls: {},
        requestIcon: function (iconName) {
            if (!(iconName in cache.urls)) {
                cache.urls[iconName] = null;
                var url = resource + "?name=" + iconName;
                $http.get(url, {responseType: 'blob'}).then(
                        function success(response) {
                            cache.urls[iconName] = window.URL.createObjectURL(response.data);
                        }
                );
            }
        }
    };

    return cache;
};
    