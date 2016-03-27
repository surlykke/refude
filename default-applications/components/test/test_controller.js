controllers.controller('ctrl', ['$scope', '$http', function($scope, $http) {
    $scope.location = window.location; 
    $scope.host = "http://localhost:7938"; 
    $scope.root = {};
    $scope.mimetype = {};
    $scope.defaultApplication = {};
    $scope.defaultApp = function(typeName, subtypeName) {
        return $scope.defaultApplication[typeName + "/" + subtypeName];
    };

    $http.get($scope.host + "/mimetypes").success(function(data) {
        $scope.root = data;
    });

    $scope.getMimetype = function(typeName, subtypeName) {
        if ($scope.mimetype[typeName + "/" + subtypeName]) {
            return;
        }
       
        var urls = getUrls($scope.root, "mimetype", {"type" : typeName, "subtype" : subtypeName});
        $http.get(urls[0]).success(function(mimetype) {
            $scope.mimetype[typeName + "/" + subtypeName] = mimetype;
            if (mimetype.defaultApplication && !$scope.defaultApplication["typeName" + "/" + "subtypeName"]) {
                var appUrls = getUrls(mimetype, "application", {"application-id": mimetype.defaultApplication});
                $http.get(appUrls[0]).success(function(application) {
                    $scope.defaultApplication[typeName + "/" + subtypeName] = application;
                });
            }
        });
    };

    var getUrls = function(resource, relation, replacements) {
        var urls = []; 
        var expandedUrls;
        if (resource.hasOwnProperty("_links") && resource._links.hasOwnProperty(relation)) {
            if (resource._links[relation] instanceof Array) {
                urls = resource._links[relation].map(function(rel) { return rel.href; });
            }
            else {
                urls = [resource._links[relation].href];
            }
        }

        expandedUrls = urls.map( function(url) { return expandUrl(url, replacements);} ); 
        return expandedUrls;
    };

    var expandUrl = function(url, replacements) {
        var newUrl = url.replace(/{(.*?)}/g, function(match, subtype) { 
            return replacements[subtype]; 
        });
        if (newUrl.indexOf("http") !== 0) {
            newUrl = $scope.host + newUrl;
        }
        return newUrl;
    };
}]);

