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
    });

    ctrl.getIcon = function(iconName, img) {
        var xhr = new XMLHttpRequest();
        xhr.open('GET', host + "theme-icon/default/" + iconName, true);
        xhr.responseType = 'blob';
        xhr.onload = function(e) {
            img.src = window.URL.createObjectURL(this.response);
        };

        xhr.send();
    };

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
                var img = document.getElementById(mimetype);
                ctrl.getIcon("hejsa", img);
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
