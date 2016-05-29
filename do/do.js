/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function doController($q, $http, $scope, $window) {
    $scope.iconCache = makeIconCache($http);
    $scope.commandList = makeCommandList($scope, $q, $http, 
                                         "http://localhost:7938/runningapplications",
                                         ["http://localhost:7938/desktopentries/commands"]);
    
    $scope.searchTerm = "";

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

    $scope.style = function(runningApp, index) {
        if (!runningApp.hasOwnProperty("geometry")) {
            return {"display" : "none"};
        }
        var geometry = runningApp.geometry;
        var selected = runningApp === $scope.commandList.selectedCommand;
        var z_index = selected ? $scope.commandList.runningApps.length : index; 
        var res = {
            "left" : convertToPx(scale*geometry.x),
            "top" : convertToPx(scale*geometry.y),
            "width" : convertToPx(scale*geometry.w),
            "height" : convertToPx(scale*geometry.h),
            "z-index" : z_index
        };
        if (selected) {
            res["opacity"] = 1;
        }
        return res;
    };

    var convertToPx = function(val) {
        var result =  "" + Math.round(val) + "px";
        return result; 
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
  
    $http.get("http://localhost:7938/display").then(function(response) { 
        calculateGeometry(response.data.geometry);
        $scope.commandList.search();
    });
    

   $scope.$watch('commandList.commands', scrollSelectedCommandIntoView, true); 
   $scope.$watch('commandList.selectedCommand', scrollSelectedCommandIntoView, true); 
};


var doModule = angular.module('do', []);

doModule.controller('doCtrl', ['$q', '$http', '$scope', '$window', doController]);

doModule.config(['$compileProvider', function ($compileProvider) {
        $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
    }]);


