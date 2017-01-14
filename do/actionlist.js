/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */
let makeActionList = function($http, listName, url, notifierUrl, callback) {
    let obj = {
        name: listName,
        list: []
    }
    
    let update = function() {
        $http.get(url).then( 
            response => {
                let listOfPromises = response.data.map(function(actionPath) {
                    return $http.get(combineUrls(url, actionPath));
                });
                
                Promise.all(listOfPromises).then(
                    responses => { 
                        obj.list = responses.map( (response) => { 
                            action = response.data;
                            action.url = response.config.url;
                            if (action.iconUrl) {
                                action.fullIconUrl = combineUrls(response.config.url, action.iconUrl);
                            }
                            else if (action.icon) {
                                action.fullIconUrl = "http://localhost:7938/icon-service/icons/icon?name=" + action.icon + "&size=32";
                            }
                            console.log("action.fullIconUrl: ", action.fullIconUrl);
                            return action;
                        });
                        callback();
                    },
                    reason => {
                        obj.list = [] 
                    }
                );
            }
        );
    };

    let evtHandler = function(evt) { 
        update(); // FIXME
    }

    let evtSource = new EventSource(notifierUrl);

    evtSource.onerror = (event) => { 
        fullList = [];
    };

    evtSource.onopen = update;

    evtSource.addEventListener("resource-updated", evtHandler);
    evtSource.addEventListener("resource-added", evtHandler);
    evtSource.addEventListener("resource-removed", evtHandler);

    return obj;

};
        
