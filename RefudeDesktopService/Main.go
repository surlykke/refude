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
	"github.com/surlykke/RefudeServices/RefudeDesktopService/doc"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/file"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/notifications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/power"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/search"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/session"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/statusnotifications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/watch"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows"
	"github.com/surlykke/RefudeServices/lib/respond"

	"github.com/surlykke/RefudeServices/lib"
)

func pathStartsWith(r *http.Request, s string) bool {
	return strings.HasPrefix(r.URL.Path, s)
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handler http.Handler = nil

	if pathStartsWith(r, "/application") || pathStartsWith(r, "/mimetype") {
		handler = applications.Handler(r)
	} else if pathStartsWith(r, "/doc") {
		handler = doc.Handler(r)
	} else if pathStartsWith(r, "/file") {
		handler = file.Handler(r)
	} else if pathStartsWith(r, "/icon") {
		handler = icons.Handler(r)
	} else if pathStartsWith(r, "/notification") {
		handler = notifications.Handler(r)
	} else if pathStartsWith(r, "/device") {
		handler = power.Handler(r)
	} else if pathStartsWith(r, "/search") {
		handler = search.Handler(r)
	} else if pathStartsWith(r, "/session") {
		handler = session.Handler(r)
	} else if pathStartsWith(r, "/item") {
		handler = statusnotifications.Handler(r)
	} else if pathStartsWith(r, "/watch") {
		handler = watch.Handler(r)
	} else if pathStartsWith(r, "/window") {
		handler = windows.Handler(r)
	}

	if handler != nil {
		handler.ServeHTTP(w, r)
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
	go watch.Run()
	go file.Run()

	go lib.Serve("org.refude.desktop-service", http.HandlerFunc(ServeHTTP))
	_ = http.ListenAndServe(":7938", http.HandlerFunc(ServeHTTP))
}
