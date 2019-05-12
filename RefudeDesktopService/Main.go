// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"net/http"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/notifications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/power"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/statusnotifications"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/applications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows"
	"github.com/surlykke/RefudeServices/lib"
)

type ResourceCollection interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request) bool
}

var repos = []ResourceCollection{
	windows.Windows,
	applications.Resources,
	notifications.Notifications,
	statusnotifications.Items,
	power.PowerResources,
	icons.Icons,
}

func serveHttp(w http.ResponseWriter, r *http.Request) {
	for _, repo := range repos {
		if repo.ServeHTTP(w, r) {
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func main() {
	go applications.Run()
	go notifications.Run()
	go power.Run()
	go statusnotifications.Run()
	go icons.Run()

	go lib.Serve("org.refude.desktop-service", http.HandlerFunc(serveHttp))
	_ = http.ListenAndServe(":7938", http.HandlerFunc(serveHttp))
}
