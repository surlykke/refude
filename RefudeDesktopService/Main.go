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

	"github.com/surlykke/RefudeServices/RefudeDesktopService/applications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/file"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/notifications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/power"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/search"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/session"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/ss_events"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/statusnotifications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows"
	"github.com/surlykke/RefudeServices/lib/respond"

	"github.com/surlykke/RefudeServices/lib"
)

type dummy struct{}

func (dummy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	var prefix = func(p string) bool {
		return strings.HasPrefix(path, p)
	}
	if prefix("/window") {
		windows.ServeHTTP(w, r)
	} else if prefix("/application") || prefix("/mimetype") {
		applications.ServeHTTP(w, r)
	} else if prefix("/notification") {
		notifications.ServeHTTP(w, r)
	} else if prefix("/device") {
		power.ServeHTTP(w, r)
	} else if prefix("/item") {
		statusnotifications.ServeHTTP(w, r)
	} else if prefix("/icon") {
		icons.ServeHTTP(w, r)
	} else if prefix("/session") {
		session.ServeHTTP(w, r)
	} else if prefix("/search") || path == "/complete" {
		search.ServeHTTP(w, r)
	} else if prefix("/file") {
		file.ServeHTTP(w, r)
	} else if path == "/events" {
		ss_events.ServeHTTP(w, r)
	} else {
		respond.NotFound(w)
	}
}

func main() {
	go windows.Run()
	go applications.Run()
	go notifications.Run()
	go power.Run()
	go statusnotifications.Run()
	go icons.Run()
	go ss_events.Run()

	var handler = dummy{}
	go lib.Serve("org.refude.desktop-service", handler)
	_ = http.ListenAndServe(":7938", handler)
}
