/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function doController($http, $scope, $window) {
    $scope.iconCache = makeIconCache($http);
    $scope.commandList = makeCommandList($http, $scope.iconCache);
    console.log("Watching:", $scope.commandList);
    
    $scope.searchTerm = "";

    $scope.search = function() {
        $scope.commandList.get(commandsSearchUrls($scope.searchTerm));
    };

    $scope.onKeyDown = function ($event) {
        console.log("keydown: ", $event);
        if ($event.code === "Escape") {
            $window.close();
        }
        if ($event.keyIdentifier === "Down" || ($event.code === "Tab" && !$event.shiftKey)) {
            $scope.commandList.selectNext();
        }
	    else if ($event.keyIdentifier === "Up" || ($event.code === "Tab" && $event.shiftKey)) {
            $scope.commandList.selectPrevious();
        }
        else if ($event.keyIdentifier === "Enter" && $scope.commandList.isSelectionValid()) {
            executeCommand($scope.commandList.selectedCommand);
        }

        scrollSelectedCommandIntoView();
    };

    $scope.running = function(command) { 
        return command.hasOwnProperty("geometry");
    };



    var commandsSearchUrls = function(searchTerm) {
        return ["http://localhost:7938/desktopentries/commands?search=" + searchTerm,
                "http://localhost:7938/runningapplications?search=" + searchTerm];
    };
  
    var executeCommand = function (command) {
        $http.post("http://localhost:7938" + command._links.execute.href).then(
            $window.close 
        );
    };

    /**
     */
    var scrollSelectedCommandIntoView = function () {
        if ($scope.commandList.selectedCommand) {
            var contentDiv = document.getElementById("contentBox");
            var commandDiv = document.getElementById($scope.commandList.selectedCommand._links.self.href);
            if (!(contentDiv && commandDiv)) {
                return;
            }

            var contentTop = contentDiv.getBoundingClientRect().top;
            var commandTop = commandDiv.getBoundingClientRect().top;
            var contentBottom = contentDiv.getBoundingClientRect().bottom;
            var commandBottom = commandDiv.getBoundingClientRect().bottom;

            var delta = null;
            if (commandTop < contentTop) {
                // So command is (partly) above content view - we move the view upwards
                delta = commandTop - contentTop - 15;
            } else if (commandBottom > contentBottom) {
                // So command is (partly) below content view - move the view downwards
                delta = commandBottom - contentBottom + 15;
            }
            if (delta) {
                contentDiv.scrollTop = contentDiv.scrollTop + delta;
            }
        }
    };
     
    var displayGeometry;

    $http.get("http://localhost:7938/display").then(function(response) { 
        displayGeometry = response.data.geometry; 
        $scope.commandList.get(["http://localhost:7938/runningapplications"]);
    });

    var drawOutline = function(command) {
        var display = document.getElementById("display");
        var selectedDiv = document.getElementById("selectedWindow"); 
        if (command && command.hasOwnProperty("geometry")) {
            display.style.display = "inline";
            selectedDiv.style.display = "inline"; 
            var contentRect = document.getElementById("contentBox").getBoundingClientRect();

            var availableWidth = Math.min(contentRect.right - contentRect.left - 100, 200);
            var availableHeight = contentRect.bottom - contentRect.top - 60;
            var scale = Math.min(availableWidth/displayGeometry.w, availableHeight/displayGeometry.h);
            var offsetX = contentRect.right - contentRect.left - availableWidth - 30;
            var offsetY = 50;
            
            placeDiv(display,     offsetX, offsetY, displayGeometry, scale);
            placeDiv(selectedDiv, offsetX, offsetY, command.geometry, scale);
        }
        else {
            display.style.display = "none";
            selectedDiv.style.display = "none";
        }
    };
  
    var placeDiv = function(div, offsetX, offsetY, geometry, scale) {
        var convertToPix = function(val) {
            var result =  "" + Math.round(val) + "px";
            console.log("convert", val, result);
            return result; 
        };
        console.log("placeDiv", offsetX, offsetY, geometry, scale);
        div.style.left = convertToPix(offsetX + geometry.x*scale);
        div.style.top = convertToPix(offsetY + geometry.y*scale);
        div.style.width = convertToPix(geometry.w*scale);
        div.style.height = convertToPix(geometry.h*scale);
    };


   $scope.$watch('commandList.selectedCommand',drawOutline, true); 
};


var doModule = angular.module('do', []);

doModule.controller('doCtrl', ['$http', '$scope', '$window', doController]);

doModule.config(['$compileProvider', function ($compileProvider) {
        $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
    }]);


