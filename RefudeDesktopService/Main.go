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
	"github.com/surlykke/RefudeServices/lib/resource"
)

func serveHttp(w http.ResponseWriter, r *http.Request) {
	var path = resource.StandardizedPath(r.URL.Path)
	var served = resource.ServeHttp(applications.ApplicationsAndMimetypes, w, r) ||
		resource.ServeHttp(notifications.Notifications, w, r) ||
		resource.ServeHttp(power.PowerResources, w, r) ||
		resource.ServeHttp(windows.Windows, w, r) ||
		resource.ServeHttp(statusnotifications.Items, w, r)

	if !served {
		switch {
		case path == "/iconthemes":
			resource.ServeCollection(w, r, icons.GetThemes())
		case path.StartsWith("/icontheme/"):
			resource.ServeResource(w, r, icons.GetTheme(path))
		case path == "/icons":
			resource.ServeCollection(w, r, icons.GetIcons())
		case path == "/icon":
			icons.ServeNamedIcon(w, r)
		case path.StartsWith("/icon/"):
			icons.ServeIcon(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
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
