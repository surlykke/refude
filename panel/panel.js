/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function panelController($scope, $interval, $http) {
    $scope.charge = 50;
    $scope.state = 3;
    $scope.stateSymbol = ['\u25CF', '\u002B', '\u2212', '\u25CB']
    $interval(function() {
        $scope.time = new Date().toLocaleTimeString();
    });
};

var panelModule = angular.module('panel', []);
panelModule.controller('panelCtrl', ['$scope', '$interval', '$http', panelController]);
panelModule.config(['$compileProvider', function ($compileProvider) {
    $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
}]);


