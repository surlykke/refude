/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function doController($http, $scope, $window) {
    $scope.iconCache = makeIconCache($http);

    $scope.commands = [];

    $scope.searchTerm = "";

    $http.get("http://localhost:7938/runningapplications").then(
        function(response) {
            console.log("Got runningapplications", response.data);
            $scope.commands = response.data.commands;

            $scope.commands.forEach(function (command) {
                $scope.iconCache.requestIcon($scope.iconUrl(command));
            });

            if ($scope.commands.length > 0) {
                selectedCommand = $scope.commands[0];
            }
        }
    );

    $scope.search = function () {
        if ($scope.searchTerm === null) {
            $scope.searchTerm = "";
        }

        if ($scope.searchTerm.trim() === "") {
            $scope.commands = [];
        } 
        else {
            var url = commandsSearchUrl()  + $scope.searchTerm;
            $http.get(url).success(function (data) {
                $scope.commands = data.commands;

                // If the last selected command is no longer there, set selected command to first in list
                if (selectedIndex() < 0 && $scope.commands.length > 0) {
                    selectedCommand = $scope.commands[0];
                }

                $scope.commands.forEach(function (command) {
                    $scope.iconCache.requestIcon($scope.iconUrl(command));
                });
            });
        }
    };

    $scope.iconUrl = function(command) {
        if (command.hasOwnProperty("Icon")) {
            return "http://localhost:7938/icons/icon?name=" + command.Icon;
        }
        else if (command._links.hasOwnProperty("icon")) {
            return "http://localhost:7938" + command._links.icon.href;
        }
        else {
            return null;
        }
    };

    $scope.selected = function(command) {
        return  selectedCommand === command;
    };

    $scope.onKeyDown = function ($event) {
        console.log("keyDown: ", $event);
        if ($event.code === "Escape") {
            $window.close();
        }
        index = selectedIndex();
        if (index > -1) {
            if ($event.keyIdentifier === "Down" && index < $scope.commands.length - 1) {
                selectedCommand = $scope.commands[index + 1];
            } 
	    else if ($event.keyIdentifier === "Up" && index > 0) {
                selectedCommand = $scope.commands[index - 1];
            } 
            else if ($event.keyIdentifier === "Enter" && selectedCommand) {
                executeCommand(selectedCommand);
            }
        }

        scrollSelectedCommandIntoView();
    };

    commandsSearchUrl = function() {
        return "http://localhost:7938/desktopentries/commands?search=";
    };
  
    executeCommand = function (command) {
        console.log("Selected ", command);
        console.log("Posting against: ", command._links.execute.href);
        $http.post("http://localhost:7938" + command._links.execute.href).then(
            $window.close 
        );
    };


    selectedCommand = null;

    selectedIndex = function () {
        return $scope.commands.findIndex(c => c === selectedCommand);
    };

    scrollSelectedCommandIntoView = function () {
        if (selectedCommand) {
            contentDiv = document.getElementById("contentBox");
            commandDiv = document.getElementById(selectedCommand._links.self.href);
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


