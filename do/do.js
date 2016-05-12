/*
* Copyright (c) 2015, 2016 Christian Surlykke
*
* This file is part of the refude project. 
* It is distributed under the GPL v2 license.
* Please refer to the LICENSE file for a copy of the license.
*/
function doController($http, $scope, $window) {
    $scope.iconUrls = {};
    $scope.commands = []; $scope.searchTerm = ""; 
    $scope.search = function() {
        if ($scope.searchTerm === null) {
            $scope.searchTerm = "";
        }
        
        if ($scope.searchTerm.trim() === "") {
            $scope.commands = [];
        }
        else {
            var url = "http://localhost:7938/desktopentries/commands?search=" + $scope.searchTerm;
            $http.get(url).success(function(data) {
                $scope.commands = data.commands;
              
                // If the last selected command is no longer there, set selected command to first in list
                if (selectedIndex() < 0 && $scope.commands.length > 0) {
                    selectedCommandId = $scope.commands[0].Id;
                }
                console.log("Set selectedCommand to " + selectedCommandId) ;

                $scope.commands.forEach(function(command) { requestIcon(command.Icon);});
            });
        }
    };

    $scope.selected = function(command) {
        return  selectedCommandId === command.Id;
    };

    $scope.onKeyDown = function($event) {
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

    $scope.selectCommand = function(commandId) {
        console.log("Selected ", commandId);
        url = "http://localhost:7938/desktopentries/commands/" + commandId;
        $http.post(url); /*
                function(data) {
                    console.log("success");
                }, 
                function(data) {
                    console.log("Could not execute command", data);
                });*/
        $window.close(); 

        /**/
    };


    selectedCommandId = null;

    // Icon cache - TODO move to directive...
    iconQueue = [];
    fetcherIsWorking = false;
    
    fetcher = function() {
        if (iconQueue.length > 0) {
            fetcherIsWorking = true;
            var iconName = iconQueue.shift();
            if (typeof $scope.iconUrls[iconName] === "undefined") {
                var url = "http://localhost:7938/icons/icon?name=" + iconName;
                $http.get(url, {responseType: 'blob'}).then(
                    function success(response){
                        $scope.iconUrls[iconName] = window.URL.createObjectURL(response.data);
                        fetcher();
                    },
                    function error(response) {
                        $scope.iconUrls[iconName] = null;
                        fetcher();
                    }
                );

            }
            else {
                fetcher();
            }
        }
        else {
            fetcherIsWorking = false;
        }
    };

    requestIcon = function(iconName) {
        iconQueue.push(iconName);
        if (!fetcherIsWorking) {
            fetcher();
        }
    };

    // ---------------- end of icon cache ------------------------

    selectedIndex = function() {
        return $scope.commands.findIndex(c =>  c.Id === selectedCommandId );
    };

    scrollSelectedCommandIntoView = function() {
        if (selectedCommandId) {
            contentDiv = document.getElementById("contentBox");
            commandDiv = document.getElementById(selectedCommandId);
            contentDivRect = contentDiv.getBoundingClientRect(); 
            commandDivRect = commandDiv.getBoundingClientRect();
            console.log("commandDivRect:", commandDivRect);
            console.log("contentDivRect:", contentDivRect);
            console.log("commandDiv.scrollTop:", commandDiv.scrollTop)
            console.log("contentDiv.scrollTop:", contentDiv.scrollTop);
            if (commandDivRect.top < contentDivRect.top) {
                contentDiv.scrollTop = contentDiv.scrollTop - (contentDivRect.top - commandDivRect.top + 15);
            }
            else if (commandDivRect.bottom > contentDivRect.bottom) {
                contentDiv.scrollTop = contentDiv.scrollTop + (commandDivRect.bottom - contentDivRect.bottom + 15);
            }
            //document.getElementById(commandId).scrollIntoView(false);
        }
    };

};


var doModule = angular.module('do', []);

doModule.controller('doCtrl', ['$http', '$scope', '$window', doController]);

doModule.config([ '$compileProvider', function ($compileProvider) {
    $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
}]);