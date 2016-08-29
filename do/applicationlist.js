/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
var makeApplicationList = function ($http, somethingChangedCallback) {
    var obj = {
        filter: function(searchTerm) {
            searchTerm = (searchTerm ? searchTerm.trim() : "").toLowerCase();            
            var tmp = [];
            if (searchTerm !== "") {
                tmp = applications.filter(app => app.name.toLowerCase().includes(searchTerm));
            }
            return tmp;
         },
       
        urlWasActivated : function(url) {
            history[url] = new Date().getTime();
            localStorage.setItem('history', JSON.stringify(history)) 
            sortApplications();
        }            
    };

    var history = {};
    var applications = [];
    var loadHistory = function() {
        try {
            history = JSON.parse(localStorage.getItem('history') || {});
        }
        catch (err) {
            history = {}
        }
        getApplications();
    };

    var getApplications = function() {
        $http.get("http://localhost:7938/desktop-service/applications").then(function(response) {
            applications = response.data.map(function(app) {
                return { 
                    name: app.Name,
                    id : app.applicationId,
                    comment: app.Comment, 
                    url : "http://localhost:7938/desktop-service/applications/" + app.applicationId,
                    iconUrl: "http://localhost:7938/icon-service/icons/icon?name=" + app.Icon
                };
            });
            sortApplications();
        });
    };
    
    var sortApplications = function() {
        applications.sort(function(app1, app2) {
            time1 = history[app1.url] || 0;
            time2 = history[app2.url] || 0;
            return time2 - time1;
        });
        somethingChangedCallback();
    };
     
    loadHistory();
    return obj;
};
