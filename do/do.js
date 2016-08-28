/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function doController($q, $http, $scope, $window) {
    const remote = require('electron').remote
    var selectedIndex = -1;

    var next = function() { 
        if ($scope.items.length > 0) {
            selectedIndex = (selectedIndex + 1) % $scope.items.length;
        } 
    };
    
    var previous = function() { 
        if ($scope.items.length > 0) { 
            selectedIndex = (selectedIndex +  $scope.items.length - 1) % $scope.items.length;
        }
    };

    $scope.searchTerm = "";
    $scope.items = [];
    $scope.windows = [];
    $scope.selectedUrl = function() {
        var url = $scope.items[selectedIndex] ? $scope.items[selectedIndex].url : undefined;
        return url;
    };

    $scope.update = function() {
        var oldUrl = $scope.selectedUrl();
        $scope.items = [];
        $scope.items = windowList.filter($scope.searchTerm);
        console.log("items:", $scope.items);
        $scope.windows = $scope.items.filter(win => !win.state.includes("Hidden"));
        console.log("windows", $scope.windows);
        applicationList.filter($scope.searchTerm).forEach(app => $scope.items.push(app));
        $scope.items.forEach(item => iconCache.requestIcon(item.iconUrl));
        selectedIndex = oldUrl ? $scope.items.findIndex(item => item.url === oldUrl) : -1;
        selectedIndex = selectedIndex === -1 ? 0 : selectedIndex;
    };
  
    $scope.select = function(url) {
        var tmp = $scope.items.findIndex(item => item.url === url);
        if (tmp > -1) {
            selectedIndex = tmp;
        }
    };

    $scope.selectAndExecute = function(url) {
        $scope.select(url);
        if ($scope.selectedUrl() === url) {
            execute(); 
        }; 
    };

    var windowList = makeWindowList($http, $scope.update);
    var applicationList = makeApplicationList($http, $scope.update);
    var iconCache = makeIconCache($http);

    var execute = function () {
        var url = $scope.selectedUrl();
        if (url) {
            var callActivated = ! $scope.items[selectedIndex].isAWindow;
            $http.post(url).then( function(response) {
                if (callActivated) {
                    applicationList.urlWasActivated(url);
                }
                $window.close();
            });
        }
    };

    var keyActions = {
        ArrowDown : next,
        ArrowUp :  previous,
        Enter : execute, 
        " " : execute,
        Escape : function() {
            remote.getCurrentWindow().hide()
        }
    };

    $scope.onKeyDown = function ($event) {
        console.log("keyDown:", event)
        if ($event.key === "Tab") {
            action = keyActions[$event.shiftKey ? "ArrowUp" : "ArrowDown"];
        }
        else {
            action = keyActions[$event.key];
        }

        if (action) action();
    };

    $scope.iconUrl = function(item) {
        return iconCache.urls[item.iconUrl] || "../../img/1x1.png";
    };
  
    $scope.itemClass  = function(item) { 
        var tmp = "item";
        if (item.url === $scope.selectedUrl()) tmp += " selected";
        return tmp;
    };

    $scope.iconClass = function(item) {
        var tmp = "itemIcon";
        if (item.isAWindow) tmp += " windowItem";
        if (item.state && item.state.includes("Hidden")) tmp += " hidden";
        return tmp;
    };

    $scope.style = function(window, index) {
        var geometry = window.geometry;
        var selected = window.url === $scope.selectedUrl();
        var z_index = selected ? $scope.windows.length : index; 
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
        if ($scope.selectedUrl()) {
            var contentDiv = document.getElementById("contentBox");
            var selectedDiv = document.getElementById($scope.selectedUrl());
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
    
    var displayGeometry = {};
    var width = 100;
    var height = 100;
    var scale = 0.1;

    var calculateGeometry = function() {
        var display = document.getElementById("disp");
        var contentRect = display.getBoundingClientRect();
        width = contentRect.right - contentRect.left - 4;
        height = contentRect.bottom - contentRect.top - 4;
        scale = Math.min(width/displayGeometry.w, height/displayGeometry.h);
    };
  
    $http.get("http://localhost:7938/wm-service/display").then(function(response) { 
        displayGeometry = response.data.geometry;
        calculateGeometry();
        angular.element($window).bind('resize', function () {
            calculateGeometry();
        });
    });


};


var doModule = angular.module('do', []);

doModule.controller('doCtrl', ['$q', '$http', '$scope', '$window', doController]);

doModule.config(['$compileProvider', function ($compileProvider) {
        $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
}]);


