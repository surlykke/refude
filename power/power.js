/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

let powerController =  function($http, $scope) {
    console.log("Into powerController"); 
    const remote = require('electron').remote
    
    $scope.actions = [];

    $scope.actionClass = function(action) {
        return action === selectedAction ? ["line", "selected"] : ["line"];
    };

    let selectNext = () => { 
        selectedAction = selectedAction.__next; 
    }
    
    let selectPrevious = () => { 
        selectedAction = selectedAction.__previous; 
    }

    let execute = function() {
        $http.post(selectedAction.url).then(function(resp) {
            remote.getCurrentWindow().close();
        });
    };

    $scope.select = (action) => { 
        selectedAction = action; 
    }

    $scope.selectAndExecute = (action) => {
        $scope.select(action);
        execute();
    };

    let selectedAction;

    let keyActions = {
        ArrowDown : selectNext,
        ArrowUp :  selectPrevious,
        Enter : execute, 
        " " : execute,
        Escape : function() {
            remote.getCurrentWindow().close()
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

    let getActions = function() {
        let url = "http://localhost:7938/power-service/actions";
        $http.get(url)
        .then(response => {
            let listOfPromises = response.data.map(function(actionPath) {
                return $http.get(combineUrls(url, actionPath));
            });
                
            Promise.all(listOfPromises).then(
                responses => { 
                    $scope.actions = responses.map(response => { 
                        action = response.data;
                        action.url = response.config.url;
                        if (action.iconUrl) {
                            action.fullIconUrl = combineUrls(response.config.url, action.iconUrl);
                        }
                        else if (action.icon) {
                            action.fullIconUrl = "http://localhost:7938/icon-service/icons/icon?name=" + action.icon + "&size=32";
                        }
                    
                        return action;
                    });
                    console.log("Actions: ", $scope.actions);
                    let previousAction = $scope.actions[$scope.actions.length - 1];
                    $scope.actions.forEach((action) => { 
                        action.__previous = previousAction;
                        previousAction.__next = action;
                        previousAction = action;
                    });
                   
                    selectedAction = $scope.actions[0];
                    $scope.$apply(); // Apparently needed - don't know why
                },
                reason => {
                    $scope.actions = [] 
                }
            );
        });
    };

    getActions();
};

let powerModule = angular.module('power', []);
console.log("Calling controller..");
powerModule.controller('powerCtrl', ['$http', '$scope', powerController]);


