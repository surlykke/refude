/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
function batteryController($scope, $timeout, $http) {

    $scope.stateStr = function() {
        return "" + $scope.charge + "% " + 
               ['\u25CF', '\u002B', '\u2212', '\u25CB', '?'][$scope.state];
    };

    var updateBatteryInfo = function(event) {
        $http.get("http://localhost:7938/power-service/devices/battery_BAT0").then(function(response) {
            $scope.charge = response.data.Percentage;
            $scope.state =  response.data.State;
            $scope.low = function() { return  $scope.state >= 2 ? 10 : 0; };
            $scope.high = function() { return $scope.state >= 2 ? 30 : 0; };
        });
    };

    var evtSource = new EventSource("http://localhost:7938/power-service/notify");
    evtSource.onerror = function(event) {
        console.log("Error:", event);
        $scope.charge = 0;
        $scope.state = 4;
    };


    evtSource.onopen = function(event) {
        console.log("open", event);
        updateBatteryInfo();
    }; 


    evtSource.addEventListener("resource-updated", function(e) {
        updateBatteryInfo();
    });
};

var batteryModule = angular.module('battery', []);
batteryModule.controller('batteryCtrl', ['$scope', '$timeout', '$http', batteryController]);
batteryModule.config(['$compileProvider', function ($compileProvider) {
    $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
}]);


