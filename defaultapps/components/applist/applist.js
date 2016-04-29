/*
* Copyright (c) 2015, 2016 Christian Surlykke
*
* This file is part of the refude project. 
* It is distributed under the GPL v2 license.
* Please refer to the LICENSE file for a copy of the license.
*/

var controllers = angular.module('controllers', []);

function applistController($scope, $http) {
    var ctrl = this;
    var host = "http://localhost:7938";
    ctrl.mimetypes = ["__others"];
    ctrl.associatedApps = {};
    ctrl.mimetypeData = {};
    ctrl.appData = {};
    ctrl.selectedApp = null;

    ctrl.selectApp = function(app) {
        ctrl.selectedApp  = ctrl.selectedApp === app ? null : app;
    };

    ctrl.setDefaultApp = function() {
        $http.patch(host + "/mimetypes/" + ctrl.mt, {"defaultApplication" : ctrl.selectedApp});
        ctrl.mt = null;
    };

    getMimetype = function(mimetype) {
        if (ctrl.mimetypes.indexOf(mimetype) > -1) {
            return;
        }
        else {
            ctrl.mimetypes.splice(-1, 0, mimetype);
            ctrl.associatedApps[mimetype] = [];
            $http.get(host + "/mimetypes/" + mimetype).success(function(mimetypeData) {
                ctrl.mimetypeData[mimetype] = mimetypeData;
                mimetypeData.associatedApplications.forEach(function(app) {
                    ctrl.associatedApps[mimetype].push(app);
                    var index = ctrl.associatedApps["__others"].indexOf(app);
                    if (index > -1) {
                        ctrl.associatedApps["__others"].splice(index, 1);
                    }
                    else {
                        getAppdata(app);
                    }
                }); 

                if (mimetype === ctrl.mt) {
                    console.log("Setting selectedApp to:", ctrl.mimetypeData[ctrl.mt].defaultApplication);
                    ctrl.selectedApp = ctrl.mimetypeData[ctrl.mt].defaultApplication;
                }

                mimetypeData.subclassOf.forEach(getMimetype);
            });
        }
    };

    getAppdata = function(appId) {
        $http.get(host + "/desktopentry/" + appId).success(function(appData) {
            ctrl.appData[appId] = appData;
        });
    };

    
    
    $scope.$watch('mt', function() {
        if (ctrl.mt) {
            $http.get(host + "/handlers").success(function(data) {
                ctrl.associatedApps["__others"] = data.fileHandlers.sort();
                ctrl.associatedApps["__others"].forEach(getAppdata);
                getMimetype(ctrl.mt);
            });
        }
    }, true);
}

(function(angular) {
    angular.module('appConfig').component( 'applist', {
        templateUrl : 'components/applist/applist.html',
        controller : applistController,
        bindings : {
            mt : '='
        }
    });
})(window.angular);
