let createResourceCollection = ($http, resourceIndexUrl, notifyUrl, callback) => { let collection = new Map();

    let getResources = () => {
        $http.get(resourceIndexUrl).then(
            response => {
                let listOfPromises = response.data.map(resourcePath => $http.get(combineUrls(resourceIndexUrl, resourcePath)));
                                        
                Promise.all(listOfPromises).then(
                    responses => {
                        responses.forEach(response => { updateResource(response.config.url, response.data);});
                        callback();
                    },
                    response => {
                        console.log(response);
                        collection.clear();
                        callback();
                    }
                );
            }
        );
    }

    let updateResource = (url, resource) => {
        resource.url = url;
        resource._actions.forEach(action => {
            action.url = url + "?action=" + action.id;
            if (action.iconUrl) {
                action.fullIconUrl = combineUrls(url, action.iconUrl);
            }
            else if (action.icon) {
                action.fullIconUrl = "http://localhost:7938/icon-service/icons/icon?name=" + action.icon + "&size=32";
            }
        });

        collection.set(url, resource);
    };
        

    let onResourceUpdated = (evt) => {
        let resourceUrl = combineUrls(notifyUrl, evt.data);
        if (resourceUrl === resourceIndexUrl) {
            getResources();
        }
        else if (collection.has(resourceUrl)) {
            $http.get(resourceUrl).then(
                response => { 
                    updateResource(resourceUrl, response.data); 
                    callback;
                },
                error => {
                    console.log(error);
                    delete collection[resourceUrl];
                    callback();
                }
            );
        }
    };


    let evtSource = new EventSource(notifyUrl);
    evtSource.onerror = (event) => { collection.clear(); };
    evtSource.onopen = getResources;
    evtSource.addEventListener("resource-added", onResourceUpdated);

    return collection;
}
