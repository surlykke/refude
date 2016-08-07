/*
* Copyright (c) 2015, 2016 Christian Surlykke
*
* This file is part of the refude project. 
* It is distributed under the GPL v2 license.
* Please refer to the LICENSE file for a copy of the license.
*/
var appconfigCtrl = function($http, $scope, $window) {

    $scope.searchTerm = "";
    $scope.filteredMimetypes = {};

    $scope.update = function() {
        $scope.filteredMimetypes = {};
        var searchTerm = $scope.searchTerm ? $scope.searchTerm.trim().toLowerCase() : "";
        mimetypes.filter(mimetype => mimetype.comment.toLowerCase().includes(searchTerm))
                 .forEach(function(mimetype){
                    $scope.filteredMimetypes[mimetype.type] = $scope.filteredMimetypes[mimetype.type] || {};
                    $scope.filteredMimetypes[mimetype.type][mimetype.subtype] = mimetype;
                  });
        
        for (var type in $scope.filteredMimetypes) {
            if (searchTerm !== "" && $scope.expandedTypes[type] !== 2 && 
                    Object.keys($scope.filteredMimetypes[type]).length <= 5) {
                openType(type, 1);
            }
            else {
                delete $scope.expandedTypes[type];
            }
        }
    };
  
    $scope.onKeyDown = function($event) {
        if ($event.key === "Escape") {
            if ($scope.currentMimetype) {
                $scope.currentMimetype = undefined;
            }
            else {
                $window.close();
            }
        }
    };

    var requestMimeIcon = function (mime) {
        if (!(mime.hasOwnProperty("iconUrl"))) {
            mime.iconUrl = null;
            var url1 = "http://localhost:7938/icon-service/icons/icon?name=" + mime.icon + "&size=50";  
            var fallbackUrl = "http://localhost:7938/icon-service/icons/icon?name=" + mime.genericIcon + "&size=50"; 
            $http.get(url1, {responseType: 'blob', headers: {'accept': 'image/png'}}).then(
                    function(response) { 
                        mime.iconUrl = window.URL.createObjectURL(response.data); 
                    }, 
                    function(response) { 
                        $http.get(fallbackUrl, {responseType: 'blob', headers: {'accept': 'image/png'}}).then(
                                function(response) { 
                                    mime.iconUrl = window.URL.createObjectURL(response.data); 
                                }); 
                            });
        }
    };
    
    $scope.expandedTypes = {};
    $scope.expandedSubtypes = {};


    $scope.toggleExpandType = function(type) {
        $scope.expandedTypes[type] ? delete $scope.expandedTypes[type] : openType(type, 2);
    };

    var openType = function(type, value) {
        $scope.expandedTypes[type] = value;
    };

    $scope.toggleExpandSubtype = function(type, subtype, $event) {
        $scope.expandedSubtypes[type] = $scope.expandedSubtypes[type] || {};
        if ($scope.expandedSubtypes[type][subtype]) {
            delete $scope.expandedSubtypes[type][subtype];
        }
        else {
            $scope.expandedSubtypes[type][subtype] = true;
            requestMimeIcon($scope.filteredMimetypes[type][subtype]);
        }
        $event.stopPropagation();
    };

    $scope.currentMimetype = undefined;
    $scope.suggestedApps = [];
    
    $scope.edit = function(type, subtype, $event) {
        if ($scope.filteredMimetypes[type] && $scope.filteredMimetypes[type][subtype]) {
            $scope.currentMimetype =  $scope.filteredMimetypes[type][subtype];
            buildSuggestedApps();
        }
        $event.stopPropagation();
    };

    var buildSuggestedApps = function() {
        $scope.suggestedApps = [];
        buildSuggestedAppsHelper($scope.currentMimetype);
        var takenApps = new Set();
        $scope.suggestedApps.forEach(function(appSuggestion) {
            appSuggestion.apps.forEach(function(app) {
                takenApps.add(app);
            });
        });
        var otherApps = [];
        Object.keys(apps).forEach(function(key) {
            if (takenApps.has(key)) return;
            var app = apps[key];
            var exec = app.Exec ? app.Exec.toLowerCase() : "";
            if (exec.includes("%f") || exec.includes("%u")) {
                otherApps.push(apps[key]);
            }
        });
        $scope.suggestedApps.push({description: "Other applications", apps: otherApps});
    };
    
    var buildSuggestedAppsHelper = function(mime) {
        var description = "Applications that handle " + mime.comment;
        if ($scope.suggestedApps.findIndex(appSuggestion => appSuggestion.description === description) < 0) {
            var appArray = [];
            mime.associatedApplications.forEach(function(appId) {
                if (apps[appId]) {
                    appArray.push(apps[appId]);
                }
            });
            $scope.suggestedApps.push({description: description, apps: appArray});
            mime.subclassOf.forEach(function(parentStr) { 
                buildSuggestedAppsHelper(mimetypes[mimetypes.findIndex(m => m.mimetype === parentStr)]);
            }); 
        }
    };
    
    $scope.select = function(applicationId) {
        var defaultApplications = $scope.currentMimetype.defaultApplications.filter(appId => appId !== applicationId);
        defaultApplications.unshift(applicationId);
        $http.patch("http://localhost:7938/desktop-service/mimetypes/" + $scope.currentMimetype.mimetype,
                    {defaultApplications: defaultApplications});
        $scope.currentMimetype = undefined;
    };

    var mimetypes = [];
    var apps = {};

    var getStuff = function() {
       $http.get("http://localhost:7938/desktop-service/applications").then(function(response) {
            response.data.forEach(function(app) {
                apps[app.applicationId] = app;
            });
            $http.get("http://localhost:7938/desktop-service/mimetypes").then(function(response) {
                mimetypes = response.data; 
                mimetypes.forEach(function(mime) {
                    mime.defaultApplication = apps[mime.defaultApplications[0]] || {Name: "none"};
                });
                console.log("Got mimetypes: ", mimetypes);
                $scope.update();
            }); 
        }); 
    };
    
    var evtSource = new EventSource("http://localhost:7938/desktop-service/notify");
    evtSource.onopen = getStuff;
    evtSource.addEventListener("resource-updated", getStuff);
    evtSource.addEventListener("resource-added", getStuff);
    evtSource.addEventListener("resource-removed", getStuff);

};

var appConfigModule = angular.module('appConfig', ['ngAria']);
appConfigModule.controller('appconfigCtrl', ['$http', '$scope', '$window', appconfigCtrl]);
appConfigModule.directive("mimetypeList", function() {
    return {
        scope: {},
        templateUrl : "mimetype-list.html"
    };
});
appConfigModule.config(['$compileProvider', function ($compileProvider) {
        $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
}]);

