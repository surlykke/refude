/*
* Copyright (c) 2015, 2016 Christian Surlykke
*
* This file is part of the refude project. 
* It is distributed under the GPL v2 license.
* Please refer to the LICENSE file for a copy of the license.
*/
function doController($http, $scope) {
    $scope.iconUrls = {};
    $scope.commands = [];
    $scope.searchTerm = "";

    var selectedCommand = null;
    var ctrl = this;
    var iconQueue = [];
    var fetcherIsWorking = false;
    
    var fetcher = function() {
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
                console.log("have it");
                fetcher();
            }
        }
        else {
            fetcherIsWorking = false;
        }
    };

    ctrl.requestIcon = function(iconName) {
        iconQueue.push(iconName);
        if (!fetcherIsWorking) {
            fetcher();
        }
    };

    ctrl.requestIcon("application-x-executable");

    ctrl.search = function() {
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
               
                if (selectedCommand) {
                    selectedCommand = $scope.commands.find(function(cmd) { return selectedCommand.Id === cmd.Id; }) || null;
                    console.log("find command returned: ", selectedCommand);
                }
                if (selectedCommand === null && $scope.commands.length > 0) {
                    selectedCommand = $scope.commands[0];
                }
                console.log("Set selectedCommand to " + selectedCommand.Id) ;
                $scope.commands.forEach(function(command, index) {
                    ctrl.requestIcon(command.Icon);
                });
            });
        }
    };

    ctrl.clazz = function(command) {
        return  selectedCommand && selectedCommand.Id === command.Id ? "selected" : "";
    };
};


var doModule = angular.module('do', []);

doModule.controller('doCtrl', doController);

doModule.config([ '$compileProvider', function ($compileProvider) {
    console.log("Correcting whitelist");
    $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
}]);