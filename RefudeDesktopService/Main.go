// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/notifications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/power"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/statusnotifications"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/applications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows"
	"github.com/surlykke/RefudeServices/lib"
	"github.com/surlykke/RefudeServices/lib/resource"
)

var resourceCollections = []resource.ResourceCollection{
	applications.ApplicationsAndMimetypes,
	notifications.Notifications,
	power.PowerResources,
	windows.Windows,
	statusnotifications.Items,
}

func serveHttp(w http.ResponseWriter, r *http.Request) {
	for _, resourceCollection := range resourceCollections {
		if resource.ServeHttp(resourceCollection, w, r) {
			return
		}
	}

	var path = string(r.URL.Path)
	switch {
	case path == "/iconthemes":
		resource.ServeCollection(w, r, icons.GetThemes())
	case strings.HasPrefix(path, "/icontheme/"):
		resource.ServeResource(w, r, icons.GetTheme(path))
	case path == "/icons":
		resource.ServeCollection(w, r, icons.GetIcons())
	case path == "/icon":
		icons.ServeNamedIcon(w, r)
	case strings.HasPrefix(path, "/icon/"):
		icons.ServeIcon(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
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
