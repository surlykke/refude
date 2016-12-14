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
    $scope.iconUrl = function(action) {
       return "http://localhost:7938/icon-service/icons/icon?name=" + action.icon + "&size=32";
    };

    $scope.actionClass = function(action) {
        return action === $scope.actions[selected] ? ["line", "selected"] : ["line"];
    };

    let selected = 0;
    let selectNext = function() {
        if ($scope.actions.length > 0) {
            selected = (selected + 1) % $scope.actions.length;
        }
    }
    
    let selectPrevious = function() {
        if ($scope.actions.length > 0) {
            selected = (selected + $scope.actions.length - 1) % $scope.actions.length;
        }
    }

    let execute = function() {
        if ($scope.actions[selected]) {
            let url = "http://localhost:7938/power-service/actions/" + $scope.actions[selected].actionId;
            $http.post(url).then(function(resp) {
                remote.getCurrentWindow().close();
            });
        }
    };

    $scope.select = function(action) {
        for (i = 0; i < $scope.actions.length; i++) {
            if ($scope.actions[i] === action) {
                selected = i;
                break;
            }
        }
    };

    $scope.selectAndExecute = function(action) {
        $scope.select(action);
        execute();
    };

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
        $http.get("http://localhost:7938/power-service/actions").then(function(resp){
            console.log("getActions got: ", resp);
            $scope.actions = resp.data;
        });
    };

    getActions();
};

console.log("Doing angular.module...");
console.log(process.env.XDG_RUNTIME_DIR);
let powerModule = angular.module('power', []);
console.log("Calling controller..");
powerModule.controller('powerCtrl', ['$http', '$scope', powerController]);


