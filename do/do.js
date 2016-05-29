/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function doController($http, $scope, $window) {
    $scope.iconCache = makeIconCache($http);
    $scope.commandList = makeCommandList($scope, $http, $scope.iconCache);
    
    $scope.searchTerm = "";

    $scope.search = function() {
        $scope.searchTerm = $scope.searchTerm ? $scope.searchTerm.trim() : ""; 
        if ($scope.searchTerm === "") {
            $scope.commandList.get(["http://localhost:7938/runningapplications"]);
        }
        else { 
            var searchUrls = [
                "http://localhost:7938/desktopentries/commands?search=" + $scope.searchTerm,
                "http://localhost:7938/runningapplications?search=" + $scope.searchTerm
            ];
            $scope.commandList.get(searchUrls);
        }
    };

    $scope.onKeyDown = function ($event) {
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

    var executeCommand = function (command) {
        $http.post("http://localhost:7938" + command._links.execute.href).then(
            $window.close 
        );
    };

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
        $scope.search();
    });
    

    var drawOutline = function(command) {
        var display = document.getElementById("disp");
        var currentWindowDiv = document.getElementById("currentWindow"); 
        if (command && command.hasOwnProperty("geometry")) {
            currentWindowDiv.style.display = "inline"; 
            var contentRect = display.getBoundingClientRect();

            var width = contentRect.right - contentRect.left;
            var height = contentRect.bottom - contentRect.top;
            var scale = Math.min(width/displayGeometry.w, height/displayGeometry.h);
            
            placeDiv(currentWindowDiv, 0, 0, command.geometry, scale);
        }
        else {
            currentWindowDiv.style.display = "none";
        }
    };
  
    var placeDiv = function(div, offsetX, offsetY, geometry, scale) {
        var convertToPix = function(val) {
            var result =  "" + Math.round(val) + "px";
            return result; 
        };
        div.style.left = convertToPix(offsetX + geometry.x*scale);
        div.style.top = convertToPix(offsetY + geometry.y*scale);
        div.style.width = convertToPix(geometry.w*scale);
        div.style.height = convertToPix(geometry.h*scale);
    };

   $scope.$watch('commandList.selectedCommand',drawOutline, true); 
   $scope.$watch('commandList.commands', scrollSelectedCommandIntoView, true); 
   $scope.$watch('commandList.selectedCommand', scrollSelectedCommandIntoView, true); 
};


var doModule = angular.module('do', []);

doModule.controller('doCtrl', ['$http', '$scope', '$window', doController]);

doModule.config(['$compileProvider', function ($compileProvider) {
        $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
    }]);


