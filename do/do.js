/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project.
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function doController($q, $http, $scope, $window, $timeout) {
    const remote = require('electron').remote;

	$scope.searchTerm = ""
	$scope.selectedUrl = undefined	
	$scope.filteredUrls = []
	$scope.action = actionUrl => $scope.resourceMap[actionUrl] || {}
	$scope.resourceMap = {}

    $scope.version = 0;
    $scope.$watch('version', () => { $scope.filterActions()});
    $scope.$watch('searchTerm', () => { $scope.filterActions()});

	let sources = []

	function Source(notifyUrl, actionListUrl) {
		this.actionListUrl = actionListUrl
		this.evtSource = new EventSource(notifyUrl)
		
		this.evtSource.onopen = () => {
			$http.get(actionListUrl).then(response => {
				$scope.resourceMap[actionListUrl] = response.data
				response.data.forEach(actionPath => { getResource(actionPath)})
			})
		}

		this.evtSource.onerror = () => {
			if ($scope.resourceMap[actionListUrl]) {
				$scope.resourceMap[actionList].length = 0
			}
		}

		this.evtSource.addEventListener("resource-added", (evt) => {
			if (evt.data.startsWith("action")) getResource(evt.data)
		})

		this.evtSource.addEventListener("resource-updated", (evt) => {
			if (evt.data.startsWith("action")) getResource(evt.data)
		})

		this.evtSource.addEventListener("resource-removed", (evt) => {
			delete $scope.resourceMap[combineUrls(notifyUrl, evt.data)]
			updateVersion()
		})

		let getResource = (resourcePath) => {
			let resourceUrl = combineUrls(notifyUrl, resourcePath)
			$http.get(resourceUrl).then(response => {
				if (response.data.IconUrl) {
					response.data.IconUrl = combineUrls(resourceUrl, response.data.IconUrl)
				}
				else if (response.data.IconName) {
					response.data.IconUrl = iconServiceUrl(response.data.IconName, 32)
				}
				console.log(resourceUrl, response.data.IconUrl)
				$scope.resourceMap[resourceUrl] = $scope.resourceMap[resourceUrl] || {}
				Object.assign($scope.resourceMap[resourceUrl], response.data)
				updateVersion()
			})
		}

	}

	$scope.filterActions = () => {
		let term = $scope.searchTerm.toLowerCase().trim();
		$scope.filteredUrls.length = 0;
		sources.forEach(source => {
			if ($scope.resourceMap[source.actionListUrl]) {
				$scope.resourceMap[source.actionListUrl].forEach(actionPath => {
					actionUrl = combineUrls(source.actionListUrl, actionPath)
					act = $scope.resourceMap[actionUrl]
					if (!act || ["Refude Do", "refudeDo"].includes(act.Name)) return
				
					if ((term === "" && act.X !== undefined) || (term !== "" && act.Name.toLowerCase().includes(term))) {
						$scope.filteredUrls.push(actionUrl)
					}
				})
			}
		})

		if ($scope.filteredUrls.indexOf($scope.selectedUrl) < 0) {
			$scope.selectedUrl = $scope.filteredUrls[0]
		}
	}


    //let updateVersion = () => { $timeout(() => {$scope.version++})}
	let updateVersion = () => { $scope.version++}
    let history = {};

    $scope.select = function(actionUrl) {
        $scope.selectedUrl = actionUrl;
    };

    $scope.selectAndExecute = function(actionUrl) {
        $scope.select(actionUrl);
        execute();
    };

    $scope.onKeyDown = function ($event) {
        if ($event.key === "Tab") {
            keyAction = keyActions[$event.shiftKey ? "ArrowUp" : "ArrowDown"];
        }
        else {
            keyAction = keyActions[$event.key];
        }

        if (keyAction) keyAction();
    };

    $scope.actionClass  = (actionUrl) => {
        _class = ["line"];
        if (actionUrl === $scope.selectedUrl) {
            _class.push("selected");
        }
        if ($scope.action(actionUrl).X != undefined) {
            _class.push("shadow")
            if ($scope.action(actionUrl).state && $scope.action(actionUrl).state.includes("_NET_WM_STATE_HIDDEN")) {
                _class.push("dimmed");
            }
        }
        return _class;
    }

    $scope.style = function(actionUrl, index) {
        return {
            "left" : "" + Math.round(scale*$scope.action(actionUrl).X) + "px",
            "top" : "" + Math.round(scale*$scope.action(actionUrl).Y) + "px",
            "width" : "" + Math.round(scale*$scope.action(actionUrl).W) + "px",
            "height" : "" + Math.round(scale*$scope.action(actionUrl).H) + "px",
            "z-index" : $scope.selectedUrl === actionUrl ? 1000 : index,
            "opacity" : $scope.selectedUrl === actionUrl ? 0.7 : 0.3
        };
    };

    let next = () => {
		let i = $scope.filteredUrls.indexOf($scope.selectedUrl)
		if (i >= 0) {
			let len = $scope.filteredUrls.length
			$scope.selectedUrl = $scope.filteredUrls[(i + 1) % len]
		}	
        scrollSelectedCommandIntoView();
    };

    let previous = () => {
       	let i = $scope.filteredUrls.indexOf($scope.selectedUrl)
		if (i >= 0) {
			let len = $scope.filteredUrls.length
			$scope.selectedUrl = $scope.filteredUrls[(i + len - 1) % len]
		} 
        scrollSelectedCommandIntoView();
    };

    let execute = function () {
        if ($scope.selectedUrl) {
			console.log("posting: ", $scope.selectedUrl)
            $http.post($scope.selectedUrl).then( response => {
                $scope.searchTerm = ""
                remote.getCurrentWindow().hide()
                history[$scope.selectedUrl] = new Date().getTime();
                localStorage.setItem('history', JSON.stringify(history))
            }).then(error => {
                console.log(error);
            });
        }
    };

    let keyActions = {
        ArrowDown : next,
        ArrowUp :  previous,
        Enter : execute,
        " " : execute,
        Escape : function() {
            remote.getCurrentWindow().hide()
        }
    };

    let scrollSelectedCommandIntoView = function () {
        if ($scope.selectedUrl) {
            let contentDiv = document.getElementById("contentBox");
            let selectedDiv = document.getElementById($scope.selectedUrl);

            if (contentDiv && selectedDiv) {
                let contentRect = contentDiv.getBoundingClientRect();
                let itemRect = selectedDiv.getBoundingClientRect();

                let delta = null;
                if (itemRect.top < contentRect.top) {
                    delta = itemRect.top - contentRect.top - 15;
                }
                else if (itemRect.bottom > contentRect.bottom) {
                    delta = itemRect.bottom - contentRect.bottom + 15;
                }

                if (delta) {
                    contentDiv.scrollTop = contentDiv.scrollTop + delta;
                }
            }
        }
    };

    let displayGeometry = {};
    let width = 100;
    let height = 100;
    let scale = 0.1;

    let calculateGeometry = function() {
		console.log("calculateGeometry, displayGeometry: ", displayGeometry)
        let display = document.getElementById("disp");
        let contentRect = display.getBoundingClientRect();
        width = contentRect.right - contentRect.left - 4;
        height = contentRect.bottom - contentRect.top - 4;
        scale = Math.min(width/displayGeometry.W, height/displayGeometry.H);
    };

    try {
        history = JSON.parse(localStorage.getItem('history') || {});
    }
    catch (err) {
        history = {}
    }

    $http.get("http://localhost:7938/wm-service/display").then(function(response) {
        displayGeometry = response.data
        calculateGeometry();
        angular.element($window).bind('resize', calculateGeometry);
    });

	sources.push(new Source("http://localhost:7938/wm-service/notify", "http://localhost:7938/wm-service/actions"))
	sources.push(new Source("http://localhost:7938/desktop-service/notify", "http://localhost:7938/desktop-service/actions"))

};


let doModule = angular.module('do', []);

doModule.controller('doCtrl', ['$q', '$http', '$scope', '$window', '$timeout', doController]);

doModule.config(['$compileProvider', function ($compileProvider) {
        $compileProvider.imgSrcSanitizationWhitelist(/^\s*((https?|ftp|file|blob|chrome-extension):|data:image\/)/);
}]);
