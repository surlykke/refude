/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
makeCommandList = function ($scope, $q, $http, runningAppsUrl, otherUrls) {
    var searchCounter = 0;

    var commandList = {
        selectedCommand: undefined,
        runningApps: [],
        commands: [],
        search: function () {
            var temp = ++searchCounter;

            var listOfPromises = [];
            if (!$scope.searchTerm || $scope.searchTerm === "") {
                listOfPromises = [$http.get(runningAppsUrl)];
            }
            else {
                listOfPromises = [$http.get(runningAppsUrl + "?search=" + $scope.searchTerm)];
                otherUrls.forEach(function(url) {
                    listOfPromises.push($http.get(url + "?search=" + $scope.searchTerm));
                });
            }
            
            $q.all(listOfPromises).then(
                    function (responses) {
                        if (temp < searchCounter) {
                            // So another search has been started - we discard results of this one
                            return;
                        }
          
                        commandList.runningApps = [];
                        var otherCommands = [];
                        commandList.commands = [];

                        responses.forEach(function (response, index) {
                            response.data.commands.forEach(function (cmd) {
                                if (index === 0) {
                                    if ("Refude Do" !== cmd.Name) {
                                        commandList.runningApps.push(cmd);
                                    }
                                } 
                                else {
                                    otherCommands.push(cmd);
                                }
                                $scope.iconCache.requestIcon($scope.iconUrlExt(cmd));
                            });
                        });

                        otherCommands.sort(function (c1, c2) {
                            return c2.lastActivated - c1.lastActivated;
                        });
                       
                        commandList.runningApps.forEach(function(cmd) {
                            commandList.commands.push(cmd);
                        });
                        otherCommands.forEach(function(cmd) {
                            commandList.commands.push(cmd);
                        });

                        if (!commandList.isSelectionValid()) {
                            commandList.selectFirst();
                        }
                    });
                },
        selectFirst: function () {
            commandList.selectedCommand = commandList.commands[0];
        },
        selectNext: function () {
            var index = commandList.commands.indexOf(commandList.selectedCommand);
            if (index >= 0 && index < commandList.commands.length - 1) {
                commandList.selectedCommand = commandList.commands[index + 1];
            }
        },
        selectPrevious: function () {
            var index = commandList.commands.indexOf(commandList.selectedCommand);
            if (index > 0) {
                commandList.selectedCommand = commandList.commands[index - 1];
            }
        },
        isSelectionValid: function () {
            return commandList.commands.indexOf(commandList.selectedCommand) > -1;
        }
    };

    return commandList;
};

