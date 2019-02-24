// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"
	"github.com/surlykke/RefudeServices/lib"
	"github.com/surlykke/RefudeServices/lib/server"
	"net/http"
	"strings"
)

var resourceServers []server.ResourceServer

func serveHttp(w http.ResponseWriter, r *http.Request) {
	for _, resourceServer := range resourceServers {
		for _, handledPrefix := range resourceServer.HandledPrefixes() {
			if strings.HasPrefix(r.URL.Path, handledPrefix) {
				switch r.Method {
				case "GET":
					resourceServer.GET(w, r)
				case "POST":
					resourceServer.POST(w, r)
				case "PATCH":
					resourceServer.PATCH(w, r)
				case "DELETE":
					resourceServer.DELETE(w, r)
				default: w.WriteHeader(http.StatusMethodNotAllowed)
				}

				return
			}
		}
	}

	w.WriteHeader(http.StatusNotFound)
}



func main() {
	/*var applicationsCollection = applications.MakeDesktopApplicationCollection()
	var mimetypeCollection = applications.MakeMimetypecollection()
	go applications.Run(applicationsCollection, mimetypeCollection)

	var windowCollection = windows.MakeWindowCollection()
	go windows.Run(windowCollection)

	var notificationCollection = notifications.MakeNotificationsCollection()
	go notifications.Run(notificationCollection)

	var powerCollection = power.MakePowerCollection()
	go power.Run(powerCollection)

	var itemCollection = statusnotifications.MakeItemCollection()
	var menuCollection = statusnotifications.MakeMenuCollection(itemCollection)
	go statusnotifications.Run(itemCollection)

	resourceServers = []server.ResourceServer{applicationsCollection, mimetypeCollection, windowCollection, notificationCollection , powerCollection, itemCollection, menuCollection}
	*/
	var iconCollection = icons.MakeIconCollection()
	icons.Run(iconCollection)
	resourceServers = []server.ResourceServer{iconCollection}
	lib.Serve("org.refude.desktop-service", http.HandlerFunc(serveHttp))
}
