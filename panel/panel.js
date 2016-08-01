/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function panelController($scope, $timeout) {
    var setclock = function() {
        var now = new Date();
        $scope.time = now.toLocaleTimeString();
        $timeout(setclock, 1000 - now.getMilliseconds() + 3); // Just after next turn of second..
    };

    setclock();
};

var panelModule = angular.module('panel', []);
panelModule.controller('panelCtrl', ['$scope', '$timeout', panelController]);


