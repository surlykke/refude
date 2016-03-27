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

    // Collect info from server
    console.log("Fetching: '" + host + "mimetypes'" );
    $http.get(host + "mimetypes").success(function(data) {
        ctrl.root = data;
        ctrl.types = Object.keys(ctrl.root.mimetypes);
        /*ctrl.types.forEach(function(type) {
            ctrl.root.mimetypes[type].forEach(function(subtype) {
                ctrl.getMimetype(type + '/' + subtype);
            });
        });*/
    });

    ctrl.getMimetype = function(mimetype) {
        var url = host + "mimetypes/" + mimetype
        console.log("Fetching '" + url + "'\n");
        $http.get(url).success(function(mimetypeObj) {
            ctrl.mimetype[mimetype] = mimetypeObj;
            if (mimetypeObj.defaultApplication) {
                var appId = mimetypeObj.defaultApplication;
                var appUrl = host + "desktopentry/" + appId;
                console.log("Fetching '" + appUrl + "'");
                $http.get(appUrl).success(function(appObj) {
                    ctrl.application[appId] = appObj;
                });
            }
        });
    };

    ctrl.selectMt = function(type, subtype) {
        ctrl.mt = type + "/" + subtype;
    };

    ctrl.evtSource = new EventSource(host + "notify");
    ctrl.evtSource.addEventListener("mimetype-updated", function(event) {
        console.log("mimetype-updated: ", event);
        console.log("data: ", event.data);
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
