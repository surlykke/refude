/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
makeCommandList = function($http, iconCache) {
    var searchCounter = 0;
    var _addCommands = function(response) {
        response.data.commands.forEach(function(command) {
            if ("Refude Do" !== command.Name) {
                commandList.commands.push(command);
                iconCache.requestIcon(commandList.iconUrl(command));
            }
        });
    };

    var commandList = {
        selectedCommand : undefined,
        commands: [], 
        get: function (listOfUrls) {
            var temp = ++searchCounter;
            var listOfPromises = listOfUrls.map(function(url) {
                return $http.get(url);
            });
            Promise.all(listOfPromises).then(
                function(responses) {
                    if (temp < searchCounter) {
                        // So another search has been started - we discard results
                        return;
                    }
                    commandList.commands = [];
                    responses.forEach(_addCommands);

                    commandList.commands.sort(function(c1, c2) { 
                        return c2.lastActivated - c1.lastActivated; 
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
        },
        iconUrl: function(command) {
            if (command.hasOwnProperty("Icon")) {
                return "http://localhost:7938/icons/icon?name=" + command.Icon;
            }
            else if (command._links.hasOwnProperty("icon")) {
                return "http://localhost:7938" + command._links.icon.href;
            }
            else {
                return null;
            }
        }
    };

    return commandList;
};
        
