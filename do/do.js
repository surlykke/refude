/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */


function doController($http, $scope, $window) {
    $scope.iconCache = makeIconCache($http, "http://localhost:7938/icons/icon");

    $scope.commands = [];

    $scope.searchTerm = "";

    $scope.search = function () {
        if ($scope.searchTerm === null) {
            $scope.searchTerm = "";
        }

        if ($scope.searchTerm.trim() === "") {
            $scope.commands = [];
        } 
        else {
            var url = "http://localhost:7938/desktopentries/commands?search=" + $scope.searchTerm;
            $http.get(url).success(function (data) {
                $scope.commands = data.commands;

                // If the last selected command is no longer there, set selected command to first in list
                if (selectedIndex() < 0 && $scope.commands.length > 0) {
                    selectedCommandId = $scope.commands[0].Id;
                }

                $scope.commands.forEach(function (command) {
                    $scope.iconCache.requestIcon(command.Icon);
                });
            });
        }
    };

    $scope.selected = function (command) {
        return  selectedCommandId === command.Id;
    };

    $scope.onKeyDown = function ($event) {
        index = selectedIndex();
        if (index > -1) {
            if ($event.keyIdentifier === "Down" && index < $scope.commands.length - 1) {
                selectedCommandId = $scope.commands[index + 1].Id;
            } 
	    else if ($event.keyIdentifier === "Up" && index > 0) {
                selectedCommandId = $scope.commands[index - 1].Id;
            } 
            else if ($event.keyIdentifier === "Enter" && selectedCommandId) {
                $scope.selectCommand(selectedCommandId);
            }
        }

        scrollSelectedCommandIntoView();
    };

    $scope.selectCommand = function (commandId) {
        console.log("Selected ", commandId);
        url = "http://localhost:7938/desktopentries/commands/" + commandId;
        $http.post(url);
        $window.close();
    };


    selectedCommandId = null;

    selectedIndex = function () {
        return $scope.commands.findIndex(c => c.Id === selectedCommandId);
    };

    scrollSelectedCommandIntoView = function () {
        if (selectedCommandId) {
            contentDiv = document.getElementById("contentBox");
            commandDiv = document.getElementById(selectedCommandId);
            if (!(contentDiv && commandDiv)) {
                return;
            }

            contentTop = contentDiv.getBoundingClientRect().top;
            commandTop = commandDiv.getBoundingClientRect().top;
            contentBottom = contentDiv.getBoundingClientRect().bottom;
            commandBottom = commandDiv.getBoundingClientRect().bottom;

            delta = null;
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
};


var doModule = angular.module('do', []);

doModule.controller('doCtrl', ['$http', '$scope', '$window', doController]);

doModule.config(['$compileProvider', function ($compileProvider) {
        $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
    }]);


