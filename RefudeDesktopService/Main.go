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
	"github.com/surlykke/RefudeServices/lib/resource"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/notifications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/power"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/statusnotifications"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/applications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows"
	"github.com/surlykke/RefudeServices/lib"
)

var repos = []resource.ResourceCollection{
	windows.Windows,
	applications.ResourceRepo,
	notifications.Notifications,
	statusnotifications.Items,
	power.PowerResources,
	icons.IconRepo,
}

func serveHttp(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	for _, repo := range repos {
		if resource := repo.Get(path); resource != nil {
			resource.ServeHttp(w, r)
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
