/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function doController($q, $http, $scope, $window) {
    const remote = require('electron').remote
    let selectedIndex = -1;

    let next = function() { 
        if ($scope.items.length > 0) {
            selectedIndex = (selectedIndex + 1) % $scope.items.length;
        } 
        scrollSelectedCommandIntoView();
    };
    
    let previous = function() { 
        if ($scope.items.length > 0) { 
            selectedIndex = (selectedIndex +  $scope.items.length - 1) % $scope.items.length;
        }
        scrollSelectedCommandIntoView();
    };

    $scope.searchTerm = "";
    $scope.items = [];
    $scope.windows = [];
    $scope.selectedUrl = function() {
        let url = $scope.items[selectedIndex] ? $scope.items[selectedIndex].url : undefined;
        return url;
    };

    $scope.update = function() {
        let oldUrl = $scope.selectedUrl();
        $scope.items = [];
        $scope.items = windowList.filter($scope.searchTerm);
        console.log("items:", $scope.items);
        $scope.windows = $scope.items.filter(win => !win.state.includes("Hidden"));
        console.log("windows", $scope.windows);
        applicationList.filter($scope.searchTerm).forEach(app => $scope.items.push(app));
        selectedIndex = oldUrl ? $scope.items.findIndex(item => item.url === oldUrl) : -1;
        selectedIndex = selectedIndex === -1 ? 0 : selectedIndex;
    };
  
    $scope.select = function(url) {
        let tmp = $scope.items.findIndex(item => item.url === url);
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

    let windowList = makeWindowList($http, $scope.update);
    let applicationList = makeApplicationList($http, $scope.update);

    let execute = function () {
        let url = $scope.selectedUrl();
        if (url) {
            let callActivated = ! $scope.items[selectedIndex].isAWindow;
            $http.post(url).then( function(response) {
                if (callActivated) {
                    applicationList.urlWasActivated(url);
                }
                $scope.searchTerm = ""
                remote.getCurrentWindow().hide()
            });
        }
    };

    let keyActions = {
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

    $scope.itemClass  = function(item) { 
        return item.url === $scope.selectedUrl() ? ["line", "selected"] : ["line"]
    };

    $scope.iconClass = function(item) {
        let tmp = "itemIcon";
        if (item.isAWindow) tmp += " windowItem";
        if (item.state && item.state.includes("Hidden")) tmp += " hidden";
        return tmp;
    };

    $scope.style = function(window, index) {
        let geometry = window.geometry;
        let selected = window.url === $scope.selectedUrl();
        let z_index = selected ? $scope.windows.length : index; 
        let res = {
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

    $scope.contentRectTop = 0;
    $scope.itemRectTop = 0;
    $scope.contentRectBottom = 0;
    $scope.itemRectBottom = 0;
    $scope.scrollDelta = 0;
    $scope.op = "";

    let scrollSelectedCommandIntoView = function () {
        if ($scope.selectedUrl()) {
            let contentDiv = document.getElementById("contentBox");
            let selectedDiv = document.getElementById($scope.selectedUrl());
            console.log("Have contentDiv and selectedDiv"); 
            $scope.contentRectTop = 0;
            $scope.itemRectTop = 0;
            $scope.contentRectBottom = 0;
            $scope.itemRectBottom = 0
            $scope.scrollDelta = 0;

            $scope.op = "";             
            if (contentDiv && selectedDiv) {
                let contentRect = contentDiv.getBoundingClientRect();
                let itemRect = selectedDiv.getBoundingClientRect();
                $scope.contentRectTop = contentRect.top;
                $scope.itemRectTop = itemRect.top;
                $scope.contentRectBottom = contentRect.bottom;
                $scope.itemRectBottom = itemRect.bottom;
                $scope.op = "";             
                console.log("contentRect:", contentRect, ", itemRect:", itemRect);
                let delta = null;
                if (itemRect.top < contentRect.top) {
                    $scope.op = "delta = " + itemRect.top + " - " + contentRect.top + " - 15";
                    delta = itemRect.top - contentRect.top - 15;
                }
                else if (itemRect.bottom > contentRect.bottom) {
                    $scope.op = "delta = " + itemRect.bottom + " - " + contentRect.bottom + " + 15";
                    delta = itemRect.bottom - contentRect.bottom + 15;
                } 
                if (delta) {
                    $scope.scrollDelta = delta;
                    contentDiv.scrollTop = contentDiv.scrollTop + delta;
                }
                else {
                    scrollDelta = 0;
                }
            }
        }
    };
    
    let displayGeometry = {};
    let width = 100;
    let height = 100;
    let scale = 0.1;

    let calculateGeometry = function() {
        let display = document.getElementById("disp");
        let contentRect = display.getBoundingClientRect();
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


let doModule = angular.module('do', []);

doModule.controller('doCtrl', ['$q', '$http', '$scope', '$window', doController]);

doModule.config(['$compileProvider', function ($compileProvider) {
        $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
}]);


