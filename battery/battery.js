/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
function batteryController($scope, $http) {
    var state = 4;

    $scope.charge = 0;
    $scope.stateStr = "";
    $scope.batteryclass = "battery-good";

    var update = function() {
        var stateChar = ["\u25CF", "\u002B", "\u2212", "\u25CB", "?"][state];
        $scope.stateStr = "" + $scope.charge + "% (" + stateChar + ")";
        
        if (state < 2 || $scope.charge > 20) {
            $scope.batteryclass = "battery-good";
        }
        else if ($scope.charge > 10) {
            $scope.batteryclass = "battery-low";
        }
        else {
            $scope.batteryclass = "battery-critical";
        }
        console.log("update done, $scope.stateStr: ", $scope.stateStr, 
                    ", $scope.batteryclass: ", $scope.batteryclass);
    };

    var updateBatteryInfo = function(event) {
        $http.get("http://localhost:7938/power-service/devices/DisplayDevice").then(function(response) {
            $scope.charge = response.data.Percentage;
            state =  response.data.State;
            console.log("charge now: ", $scope.charge, ", state:", state);
            update(); 
        });
    };

    var evtSource = new EventSource("http://localhost:7938/power-service/notify");
    evtSource.onerror = function(event) {
        console.log("Error:", event);
        $scope.charge = 0;
        state = 4;
        update();
    };

    evtSource.onopen = function(event) {
        console.log("open", event);
        updateBatteryInfo();
    }; 

    evtSource.addEventListener("resource-updated", function(e) {
        updateBatteryInfo();
    });

    var openPowerSettings = function() {
        $http.post("http://localhost:7938/desktop-service/applications/lxqt-leave.desktop");
    };

    chrome.contextMenus.create({"id": "PowerSettings", "title": "Power settings"});

    chrome.contextMenus.onClicked.addListener(openPowerSettings);
};

var batteryModule = angular.module('battery', []);
batteryModule.controller('batteryCtrl', ['$scope', '$http', batteryController]);
