// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/surlykke/RefudeServices/RefudeDesktopService/applications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/notifications"
	"github.com/surlykke/RefudeServices/lib"
	"net/http"
	"strings"
)


func server(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/application") {
		applications.ApplicationsServer.ServeHTTP(w,r)
	} else if strings.HasPrefix(r.URL.Path, "/mimetype") {
		applications.MimetypesServer.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	//var resourceMap = resource.MakeJsonResourceMap()

	/*go applications.Run(resourceMap)
	go windows.Run(resourceMap)
	go power.Run(resourceMap)*/
	go notifications.Run(resourceMap);
	//go statusnotifications.Run(resourceMap)
	go applications.Run()
	lib.Serve("org.refude.desktop-service", http.HandlerFunc(server))
}
