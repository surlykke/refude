// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/surlykke/RefudeServices/RefudeDesktopService/applications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/notifications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/power"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/statusnotifications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows"
	"github.com/surlykke/RefudeServices/lib"
	"net/http"
)

func serveHttp(w http.ResponseWriter, r *http.Request) {
	if ! (applications.Serve(w, r) || icons.Serve(w, r) || notifications.Serve(w, r) ||
		power.Serve(w, r) || statusnotifications.Serve(w, r) || windows.Serve(w, r)) {

		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	go applications.Run()
	go windows.Run()
	go notifications.Run()
	go power.Run()
	go statusnotifications.Run()
	go icons.Run()

	go lib.Serve("org.refude.desktop-service", http.HandlerFunc(serveHttp))
	http.ListenAndServe(":7938", http.HandlerFunc(serveHttp))
}
