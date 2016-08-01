/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function doController($q, $http, $scope, $window) {
    $scope.iconCache = makeIconCache($http);
    $scope.itemList = makeItemList($http, $scope.iconCache); 
    $scope.searchTerm = "";

    var execute = function () {
        var url = $scope.itemList.selectedUrl();
        if (url) {
            $http.post(url).then( function(response) {
                $window.close();
            });
        }
    };

    var keyActions = {
        ArrowDown : $scope.itemList.next,
        ArrowUp :  $scope.itemList.previous,
        Enter : execute, 
        " " : execute
    };

    $scope.onKeyDown = function ($event) {
        if ($event.key === "Tab") {
            action = keyActions[$event.shiftKey ? "ArrowUp" : "ArrowDown"];
        }
        else {
            action = keyActions[$event.key];
        }

        if (action) action();
    };

    $scope.iconUrl = function(item) {
        return $scope.iconCache.urls[item.iconUrl] || "../../img/1x1.png";
    };
   
    $scope.windowClass = function(item) {
        return item.isAWindow ? "windowItem" : "";
    };

    $scope.selectedClass  = function(item) {
        return item.url === $scope.itemList.selectedUrl() ? "selected" : "";
    };

    $scope.style = function(window, index) {
        var geometry = window.geometry;
        var selected = window.url === $scope.itemList.selectedUrl();
        var z_index = selected ? $scope.itemList.filteredWindows.length : index; 
        var res = {
            "left" : "" + Math.round(scale*geometry.x) + "px",
            "top" : "" + Math.round(scale*geometry.y) + "px",
            "width" : "" + Math.round(scale*geometry.w) + "px",
            "height" : "" + Math.round(scale*geometry.h) + "px",
            "z-index" : z_index
        };
        if (selected) {
            res["opacity"] = 1;
        }
        return res;
    };

    var scrollSelectedCommandIntoView = function () {
        if ($scope.itemList.selectedUrl()) {
            var contentDiv = document.getElementById("contentBox");
            var selectedDiv = document.getElementById($scope.itemList.selectedUrl());
            if (contentDiv && selectedDiv) {
                var contentRect = contentDiv.getBoundingClientRect();
                var itemRect = selectedDiv.getBoundingClientRect();
                var delta = null;
                if (itemRect.top < contentRect.top) {
                    delta = itemRect.top - contentRect.top - 15;
                }
                else if (itemRect.bottom > contentRect.bottom) {
                    delta = itemRect.bottom - contentRect.bottom + 15;
                } 
                if (delta) {
                    contentDiv.scrollTop = contentDiv.scrollTop + delta;
                }
            }
        }
    };
     
    var width = 100;
    var height = 100;
    var scale = 0.1;

    var calculateGeometry = function(displayGeometry) {
        var display = document.getElementById("disp");
        var contentRect = display.getBoundingClientRect();
        width = contentRect.right - contentRect.left - 4;
        height = contentRect.bottom - contentRect.top - 4;
        scale = Math.min(width/displayGeometry.w, height/displayGeometry.h);
    };
  
    $http.get("http://localhost:7938/windowmanager-service/display").then(function(response) { 
        calculateGeometry(response.data.geometry);
    });

    console.log("chrome.storage.local", chrome.storage.local);
};


var doModule = angular.module('do', []);

doModule.controller('doCtrl', ['$q', '$http', '$scope', '$window', doController]);

doModule.config(['$compileProvider', function ($compileProvider) {
        $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
}]);


