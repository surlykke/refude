/*
* Copyright (c) 2015, 2016 Christian Surlykke
*
* This file is part of the refude project. 
* It is distributed under the GPL v2 license.
* Please refer to the LICENSE file for a copy of the license.
*/

function mimetypelistController($http) {
    var ctrl = this;

    var host = "http://localhost:7938/"; 
    // root contains a map topleveltype -> list of subtypes for that topleveltype 
    ctrl.root = {};
   
    // Contains a list of toplevel types
    ctrl.types = [];

    // Contains a map from mimetype as string (eg. 'text/html') to an object describing that type
    ctrl.mimetype = {};

    // --
    ctrl.application = {};

    ctrl.getMimetypes = function() {
        var url = host + "mimetypes";
        if (ctrl.searchTerm && "" !== ctrl.searchTerm) {
            url = url + "?search=" + ctrl.searchTerm;
        }
        $http.get(url).success(function(data) {
            ctrl.root = data;
            ctrl.types = Object.keys(ctrl.root.mimetypes);
        });     
    };

    // Collect info from server
    ctrl.getMimetypes();
    

    ctrl.getIcon = function(iconNames, img) {
        var xhr = new XMLHttpRequest();
        var url = host + "icons/icon?" + iconNames.map(function(iconName) { return "name=" +iconName; }).join('&');
        url = url+"&size=72";
        xhr.open('GET', url, true);
        xhr.responseType = 'blob';
        xhr.onload = function(e) {
            img.src = window.URL.createObjectURL(this.response);
        };

        xhr.send();
    };

    ctrl.getMimetype = function(mimetype) {
        var url = host + "mimetypes/" + mimetype
        $http.get(url).success(function(mimetypeObj) {
            ctrl.mimetype[mimetype] = mimetypeObj;
            var appId = mimetypeObj.defaultApplications[0];
            if (appId) {
                var appUrl = host + "desktopentries/" + appId;
                $http.get(appUrl).success(function(appObj) {
                    ctrl.application[appId] = appObj;
                });
            }           
            var img = document.getElementById(mimetype);
            ctrl.getIcon([mimetypeObj.icon, mimetypeObj.genericIcon], img);
         });
    };

    ctrl.selectMt = function(type, subtype) {
        ctrl.mt = type + "/" + subtype;
    };

    ctrl.evtSource = new EventSource(host + "notify");
    ctrl.evtSource.addEventListener("mimetype-updated", function(event) {
        ctrl.mimetype[event.data] = null;
        ctrl.getMimetype(event.data);
    });
};

angular.module('appConfig').component( 'mimetypelist', {
    templateUrl : 'components/mimetypelist/mimetypelist.html',
    controller : mimetypelistController,
    bindings: {
        mt: '='
    }
});
