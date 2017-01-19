/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function doController($q, $http, $scope, $window) {
    const remote = require('electron').remote; 

    $scope.searchTerm = "";
    $scope.actions = [];

    $scope.update = function() {
        $scope.actions = [];
        [].push.apply($scope.actions, collectWindowActions()); 
        [].push.apply($scope.actions, collectOtherActions());

        if ($scope.actions.length > 0) {
            let previous = $scope.actions[$scope.actions.length - 1];
            $scope.actions.forEach((action) => {
                action.__previous = previous;
                previous.__next = action;
                previous = action;
            });
        }
        
        if ($scope.selectedAction) {
            $scope.selectedAction = $scope.actions.find((action) => {action.url === $scope.selectedAction.url});
        } 
        
        if (!$scope.selectedAction && $scope.actions.length > 0) {
            $scope.selectedAction = $scope.actions[0];
        }

    };

    $scope.select = function(action) {
        $scope.selectedAction = action;
    };

    $scope.selectAndExecute = function(action) {
        $scope.select(action);
        execute(); 
    };

    $scope.onKeyDown = function ($event) {
        if ($event.key === "Tab") {
            action = keyActions[$event.shiftKey ? "ArrowUp" : "ArrowDown"];
        }
        else {
            action = keyActions[$event.key];
        }

        if (action) action();
    };

    $scope.actionClass  = function(action) { 
        _class = ["line"];
        if (action === $scope.selectedAction) {
            _class.push("selected");
        }
        if (action.geometry) {
            _class.push("shadow")
            if (action.state && action.state.includes("Hidden")) {
                _class.push("dimmed");
            }
        }
        return _class;
    };

    $scope.style = function(action, index) {
        return {
            "left" : "" + Math.round(scale*action.geometry.x) + "px",
            "top" : "" + Math.round(scale*action.geometry.y) + "px",
            "width" : "" + Math.round(scale*action.geometry.w) + "px",
            "height" : "" + Math.round(scale*action.geometry.h) + "px",
            "z-index" : $scope.selectedAction === action ? 1000 : index,
            "opacity" : $scope.selectedAction === action ? 0.7 : 0.3 
        };
    };

    let windowResources = createResourceCollection($http, 
                                                   "http://localhost:7938/wm-service/windows", 
                                                   "http://localhost:7938/wm-service/notify",
                                                   $scope.update);

    let otherResources = createResourceCollection($http,
                                                  "http://localhost:7938/desktop-service/applications", 
                                                  "http://localhost:7938/desktop-service/notify", 
                                                  $scope.update);

    let collectWindowActions = () => {
        let term = $scope.searchTerm.toLowerCase().trim();
        let actions = [];
        windowResources.forEach(resource => {
            if (!resource.state.includes("Above") &&
                resource.name !== "Refude Do" &&
                resource._actions[0].name.toLowerCase().includes(term)) {
                resource._actions[0].geometry = resource.geometry;
                actions.push(resource._actions[0]);
            }
        });
        return actions;
    };

    let collectOtherActions = () => {
        let term = $scope.searchTerm.toLowerCase().trim();
        if ("" === term) {
            return [];
        }
        else {
            let actions = [];
            otherResources.forEach((resource) => {
                [].push.apply(actions, resource._actions.filter(action => action.name.toLowerCase().includes(term)));
            });
            return actions;
        }
    };


    let next = function() { 
        $scope.selectedAction = $scope.selectedAction ? $scope.selectedAction.__next : undefined;
        scrollSelectedCommandIntoView();
    };
    
    let previous = function() { 
        $scope.selectedAction = $scope.selectedAction ? $scope.selectedAction.__previous : undefined;
        scrollSelectedCommandIntoView();
    };

    let includeAction = function(action) {
        let searchTerm = $scope.searchTerm.trim();
        result = (action.geometry || "" !== searchTerm) && action.name.toLowerCase().includes($scope.searchTerm.trim());  
        return result; 
    };


    let execute = function () {
        console.log("Execute, selectedAction:", $scope.selectedAction);
        let url = $scope.selectedAction.url;
        if (url) {
            console.log("Posting: ", url);
            $http.post(url).then( function(response) {
                $scope.searchTerm = ""
                remote.getCurrentWindow().hide()
            });
        }
    };

    let keyActions = {
        ArrowDown : next,
        ArrowUp :  previous,
        Enter : execute, 
        " " : execute,
        Escape : function() {
            remote.getCurrentWindow().hide()
        }
    };

    let scrollSelectedCommandIntoView = function () {
        if ($scope.selectedAction) {
            let contentDiv = document.getElementById("contentBox");
            let selectedDiv = document.getElementById($scope.selectedAction.url);

            if (contentDiv && selectedDiv) {
                let contentRect = contentDiv.getBoundingClientRect();
                let itemRect = selectedDiv.getBoundingClientRect();
                
                let delta = null;
                if (itemRect.top < contentRect.top) {
                    delta = itemRect.top - contentRect.top - 15;
                }
                else if (itemRect.bottom > contentRect.bottom) {
                    delta = itemRect.bottom - contentRect.bottom + 15;
                } 

                if (delta) {
                    contentDiv.scrollTop = contentDiv.scrollTop + delta;
                }
            }
        }
    };
    
    let displayGeometry = {};
    let width = 100;
    let height = 100;
    let scale = 0.1;

    let calculateGeometry = function() {
        let display = document.getElementById("disp");
        let contentRect = display.getBoundingClientRect();
        width = contentRect.right - contentRect.left - 4;
        height = contentRect.bottom - contentRect.top - 4;
        scale = Math.min(width/displayGeometry.w, height/displayGeometry.h);
    };
  
    $http.get("http://localhost:7938/wm-service/display").then(function(response) { 
        displayGeometry = response.data.geometry;
        calculateGeometry();
        angular.element($window).bind('resize', calculateGeometry); 
    });
};


let doModule = angular.module('do', []);

doModule.controller('doCtrl', ['$q', '$http', '$scope', '$window', doController]);

doModule.config(['$compileProvider', function ($compileProvider) {
        $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
}]);


