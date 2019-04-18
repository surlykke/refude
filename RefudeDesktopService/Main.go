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
	switch {
	case path == "/windows":
		resource.ServeCollection(w, r, windows.GetWindows())
	case path.StartsWith("/window/"):
		resource.ServeResource(w, r, windows.GetWindow(path))
	case path == "/applications":
		resource.ServeCollection(w, r, applications.GetApplications())
	case path.StartsWith("/application/"):
		resource.ServeResource(w, r, applications.GetApplication(path))
	case path == "/notifications":
		resource.ServeCollection(w, r, notifications.GetNotifications())
	case path.StartsWith("/notification/"):
		resource.ServeResource(w, r, notifications.GetNotification(path))
	case path == "/devices":
		resource.ServeCollection(w, r, power.GetDevices())
	case path.StartsWith("/device/"):
		resource.ServeResource(w, r, power.GetDevice(path))
	case path == "/session":
		resource.ServeResource(w, r, power.Session)
	case path == "/items":
		resource.ServeCollection(w, r, statusnotifications.GetItems())
	case path.StartsWith("/item/"):
		resource.ServeResource(w, r, statusnotifications.GetItem(path))
	case path.StartsWith("/itemmenu/"):
		resource.ServeResource(w, r, statusnotifications.GetMenu(path))
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

func main() {
	go applications.Run()
	go notifications.Run()
	go power.Run()
	go statusnotifications.Run()
	go icons.Run()

	go lib.Serve("org.refude.desktop-service", http.HandlerFunc(serveHttp))
	_ = http.ListenAndServe(":7938", http.HandlerFunc(serveHttp))
}
