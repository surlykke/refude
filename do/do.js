/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function doController($q, $http, $scope, $window, $timeout) {
    const remote = require('electron').remote; 

    $scope.searchTerm = "";
    $scope.actions = [];
    $scope.version = 0;

    let updateVersion = () => { $timeout(() => {$scope.version++;});};
    let windowResourceFilter = res => !(res.state.includes("Above") || ["Refude Do", "refudeDo"].includes(res.name));
     
    let windowResourceActions = createResourceCollection($http, 
                                                         "http://localhost:7938/wm-service/windows", 
                                                         "http://localhost:7938/wm-service/notify",
                                                         windowResourceFilter,
                                                         updateVersion);

    let applicationResourceFilter = res => true; 
    let applicationResourceActions = createResourceCollection($http, 
                                                              "http://localhost:7938/desktop-service/applications", 
                                                              "http://localhost:7938/desktop-service/notify",
                                                              applicationResourceFilter,
                                                              updateVersion);
    
    let filterActions = () => {
        $scope.actions.length = 0;
        let term = $scope.searchTerm.toLowerCase().trim();
        let windowActionFilter = (action) => term.length === 0 || action.name.toLowerCase().includes(term);
        windowResourceActions.filter(windowActionFilter).forEach(action => $scope.actions.push(action));
        let applicationActionFilter = (action) => term.length > 0 && action.name.toLowerCase().includes(term);
        applicationResourceActions.filter(applicationActionFilter).forEach(action => $scope.actions.push(action));

        if ($scope.actions.length > 0) {
            let previous = $scope.actions[$scope.actions.length - 1];
            $scope.actions.forEach((action) => {
                action._previous = previous;
                previous._next = action;
                previous = action;
            });
        }
      
        if ($scope.selectedAction) {
            $scope.selectedAction = $scope.actions.find((action) => action.url === $scope.selectedAction.url);
        }
            
        if (!$scope.selectedAction && $scope.actions.length > 0) {
            $scope.selectedAction = $scope.actions[0];
        } 

    }


    $scope.$watch('version', filterActions);
    $scope.$watch('searchTerm', filterActions);

    
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
        if (action.resource.geometry) {
            _class.push("shadow")
            if (action.state && action.state.includes("Hidden")) {
                _class.push("dimmed");
            }
        }
        return _class;
    };

    $scope.style = function(action, index) {
        return {
            "left" : "" + Math.round(scale*action.resource.geometry.x) + "px",
            "top" : "" + Math.round(scale*action.resource.geometry.y) + "px",
            "width" : "" + Math.round(scale*action.resource.geometry.w) + "px",
            "height" : "" + Math.round(scale*action.resource.geometry.h) + "px",
            "z-index" : $scope.selectedAction === action ? 1000 : index,
            "opacity" : $scope.selectedAction === action ? 0.7 : 0.3 
        };
    };

    let next = () => { 
        if ($scope.selectedAction) {
            $scope.selectedAction = $scope.selectedAction._next;
        } 
        scrollSelectedCommandIntoView();
    };
    
    let previous = () => { 
        if ($scope.selectedAction) { 
            $scope.selectedAction = $scope.selectedAction._previous;
        }
        scrollSelectedCommandIntoView();
    };

    let execute = function () {
        let url = $scope.selectedAction.url;
        if (url) {
            $http.post(url).then( response => {
                $scope.searchTerm = ""
                remote.getCurrentWindow().hide()
            }).then(error => {
                console.log(error);
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

doModule.controller('doCtrl', ['$q', '$http', '$scope', '$window', '$timeout', doController]);

doModule.config(['$compileProvider', function ($compileProvider) {
        $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
}]);


