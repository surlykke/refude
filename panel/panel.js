/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function panelController($scope, $timeout, $http) {
    $scope.stateSymbol = ['\u25CF', '\u002B', '\u2212', '\u25CB', '?'];

    var updateBatteryInfo = function(event) {
        $http.get("http://localhost:8080/power/devices/battery_BAT0").then(function(response) {
            $scope.charge = response.data.Percentage;
            $scope.state = response.data.State;
            $scope.low = function() { return  $scope.state >= 2 ? 10 : 0; };
            $scope.high = function() { return $scope.state >= 2 ? 30 : 0; };
        });
    };

    var evtSource = new EventSource("http://localhost:8080/power/notify");
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

    var setclock = function() {
        var now = new Date();
        $scope.time = now.toLocaleTimeString();
        $timeout(setclock, 1000 - now.getMilliseconds() + 3); // Just after next turn of second..
    };

    setclock();
};

var panelModule = angular.module('panel', []);
panelModule.controller('panelCtrl', ['$scope', '$timeout', '$http', panelController]);
panelModule.config(['$compileProvider', function ($compileProvider) {
    $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
}]);


