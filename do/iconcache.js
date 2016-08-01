/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

makeIconCache = function ($http) {

    var cache = {
        urls: {},
        requestIcon: function (url) {
            if (!url) {
                return;
            }
            if (!(url in cache.urls)) {
                cache.urls[url] = null;
                $http.get(url, {responseType: 'blob', headers: {'accept': 'image/png'}}).then(
                    function success(response) {
                        cache.urls[url] = window.URL.createObjectURL(response.data);
                    }
                );
            }
        }
    };

    return cache;
};
    