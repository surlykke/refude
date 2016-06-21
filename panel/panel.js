/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function panelController($scope, $interval, $http) {
    $scope.stateSymbol = ['\u25CF', '\u002B', '\u2212', '\u25CB', '?'];

    var updateBatteryInfo = function(event) {
        if (event.type === "error") {
            $scope.charge = 0;
            $scope.state = 4;
        }
        else {
            $http.get("http://localhost:7939/devices/battery_BAT0").then(function(response) {
                $scope.charge = response.data.Percentage;
                $scope.state = response.data.State;
                $scope.low = function() { return  $scope.state >= 2 ? 10 : 0; };
                $scope.high = function() { return $scope.state >= 2 ? 30 : 0; };
            });
        }
    };

    var evtSource = new EventSource("http://localhost:7939/notify");
    evtSource.onmessage = evtSource.onerror = evtSource.onopen = updateBatteryInfo; 

    $interval(function() { 
        $scope.time = new Date().toLocaleTimeString();
    });

};

var panelModule = angular.module('panel', []);
panelModule.controller('panelCtrl', ['$scope', '$interval', '$http', panelController]);
panelModule.config(['$compileProvider', function ($compileProvider) {
    $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
}]);


