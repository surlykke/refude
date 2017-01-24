let createResourceCollection = ($http, resourceIndexUrl, notifyUrl, resourceFilter, callBack) => { 
    
    let actions = [];

    let resources = new Map();

    let getResources = () => {
        $http.get(resourceIndexUrl).then(
            response => {
                let listOfPromises = response.data.map(resourcePath => $http.get(combineUrls(resourceIndexUrl, resourcePath)));
                                        
                Promise.all(listOfPromises).then(
                    responses => {
                        resources.clear();
                        responses.forEach(response => { updateResource(response.config.url, response.data);});
                        updateActions();
                    },
                    response => {
                        resources.clear();
                        console.log(response);
                        updateActions();
                    }
                );
            }
        );
    }

    let updateResource = (url, resource) => {
        resource.url = url;
        resources.set(url, resource);
    };
        
    let updateActions = () => {
        actions.length = 0;
        for (let [url, resource] of resources) {
            if (resourceFilter(resource)) {
                for(id in resource._actions) {
                    let act = resource._actions[id];
                    actions.push({
                        name: act.name,
                        comment: act.comment,
                        url: resource.url + "?action=" + id,
                        iconUrl: act.iconUrl ? combineUrls(resource.url, act.iconUrl) : 
                                                act.icon ? iconServiceUrl(act.icon) : undefined,
                        resource: resource
                    });
                }
            }
        }
        callBack();
    };



    let onResourceUpdated = (evt) => {
        let resourceUrl = combineUrls(notifyUrl, evt.data);
        if (resourceUrl === resourceIndexUrl) {
            getResources();
        }
        else if (resources.has(resourceUrl)) {
            $http.get(resourceUrl).then(
                response => { 
                    updateResource(resourceUrl, response.data); 
                    updateActions();
                },
                error => {
                    console.log(error);
                    delete resources[resourceUrl];
                    updateActions();
                }
            );
        }
    };


    let evtSource = new EventSource(notifyUrl);
    evtSource.onerror = (event) => { resources.clear(); };
    evtSource.onopen = getResources;
    evtSource.addEventListener("resource-updated", onResourceUpdated);

    return actions;
}
