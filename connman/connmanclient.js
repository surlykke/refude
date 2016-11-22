/*
 * Copyright (c) 2015, 2016 Christian Surlykke
 *
 * This file is part of the refude project. 
 * It is distributed under the GPL v2 license.
 * Please refer to the LICENSE file for a copy of the license.
 */

function connmanclientController($http, $scope) {

	$scope.technologies = [];
	$scope.services = [];

	$scope.techdisplayname = function(json) {
		return json["Name"];
	}

	$scope.servicedisplayname = function(json) {
		let name = json["Name"] && json["Name"].length > 0 ?  json["Name"] : "(hidden)";
		if (json["State"] === "online") {
			name += " \u2713";
		}	
		else if (json["State"] === "association") {
			name +=  " a";
		}
		else if (json["State"] === "configuration") {
			name += " c";
		}
		return name;
	}
	
	$scope.serviceclass = function(json) {
		return ["ready", "online"].includes(json["State"]) ? ["service", "bold"] : ["service"];
	}

	$scope.technologyclass = function(technologyJson) {
		return 0 === technologyJson["Powered"] ? 
			["tech", "dimmed"] : 
			0 === technologyJson["Connected"] ? 
				["tech"] : ["tech", "connected"];
	}

   	let getTechnologies = function() {
		$http.get("http://localhost:7938/connman-service/technologies").then(function(response) {
			$scope.technologies = response.data;
		});
	};
	
   	let getServices = function() {
		$http.get("http://localhost:7938/connman-service/services").then(function(response) {
			$scope.services = response.data;
		});
	};
		
	let evtSource = new EventSource("http://localhost:7938/connman-service/notify");

    let eventHandler = function(event) {
		getTechnologies();
		getServices();
    };

    evtSource.onerror = function(event) {
		$scope.technologies = [];
		$scope.services = [];
    };

    evtSource.onopen = function(event) {
		getTechnologies();
		getServices();
    }; 

    evtSource.addEventListener("resource-updated", eventHandler);

};


let connmanclientModule = angular.module('connmanclient', []);
connmanclientModule.controller('connmanclientController', ['$http', '$scope', connmanclientController]);

