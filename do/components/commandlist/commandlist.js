/*
* Copyright (c) 2015, 2016 Christian Surlykke
*
* This file is part of the refude project. 
* It is distributed under the GPL v2 license.
* Please refer to the LICENSE file for a copy of the license.
*/

function commandlistController($http) {
    var ctrl = this;

    ctrl.commands = [];
    ctrl.searchTerm = "";
    ctrl.search = function() {
        if (ctrl.searchTerm === null) {
            ctrl.searchTerm = "";
        }
        else if (ctrl.searchTerm.trim() === "") {
            ctrl.commands = [];
        }
        else {
            var url = "http://localhost:7938/desktopentries/commands?search=" + ctrl.searchTerm;
            $http.get(url).success(function(data) {
                ctrl.commands = data.commands;
            });
        }
    }
};

angular.module('do').component( 'commandlist', {
    templateUrl : 'components/commandlist/commandlist.html',
    controller : commandlistController
});
