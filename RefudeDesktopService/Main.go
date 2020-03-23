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

	"github.com/surlykke/RefudeServices/RefudeDesktopService/sse_events"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/search"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/session"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/applications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/backlight"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/notifications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/power"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/statusnotifications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows"
	"github.com/surlykke/RefudeServices/lib/respond"

	"github.com/surlykke/RefudeServices/lib"
)

type dummy struct{}

func (dummy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	switch {
	case match(r, "/window"):
		windows.ServeHTTP(w, r)
	case match(r, "/application"):
		applications.AppServeHTTP(w, r)
	case match(r, "/mimetype"):
		applications.MimetypeServeHTTP(w, r)
	case match(r, "/notification"):
		notifications.ServeHTTP(w, r)
	case match(r, "/device"):
		power.ServeHTTP(w, r)
	case match(r, "/item"):
		statusnotifications.ServeHTTP(w, r)
	case match(r, "/backlight"):
		backlight.ServeHTTP(w, r)
	case match(r, "/icon"):
		icons.ServeHTTP(w, r)
	case match(r, "/session"):
		session.ServeHTTP(w, r)
	case match(r, "/search"):
		search.ServeHTTP(w, r)
	case path == "/complete":
		search.ServeHTTP(w, r)
	case path == "/events":
		sse_events.ServeHTTP(w, r)
	default:
		respond.NotFound(w)
	}
}

func match(r *http.Request, prefix string) bool {
	return strings.HasPrefix(r.URL.Path, prefix)
}

func main() {
	go windows.Run()
	go applications.Run()
	go notifications.Run()
	go power.Run()
	go statusnotifications.Run()
	go backlight.Run()
	go icons.Run()
	go sse_events.Run()

	var handler = dummy{}
	go lib.Serve("org.refude.desktop-service", handler)
	_ = http.ListenAndServe(":7938", handler)
}
