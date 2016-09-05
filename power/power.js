/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

let powerController =  function($http, $scope) {
    console.log("Into powerController"); 
    $scope.actions = [];
    $scope.iconUrl = function(action) {
       return "http://localhost:7938/icon-service/icons/icon?name=" + action.icon + "&size=32";
    };

    let getActions = function() {
        $http.get("http://localhost:7938/power-service/actions").then(function(resp){
            console.log("getActions got: ", resp);
            $scope.actions = resp.data;
        });
    }
    


    getActions();
};

console.log("Doing angular.module...");

let powerModule = angular.module('power', []);
console.log("Calling controller..");
powerModule.controller('powerCtrl', ['$http', '$scope', powerController]);


