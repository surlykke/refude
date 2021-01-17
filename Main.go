// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/doc"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/search"
	"github.com/surlykke/RefudeServices/session"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/windows"

	"github.com/surlykke/RefudeServices/lib"

	_ "net/http/pprof"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handler http.Handler = nil
	if strings.HasPrefix(r.URL.Path, "/application") || strings.HasPrefix(r.URL.Path, "/mimetype") {
		handler = applications.Handler(r)
	} else if strings.HasPrefix(r.URL.Path, "/doc") {
		handler = doc.Handler(r)
	} else if strings.HasPrefix(r.URL.Path, "/file") {
		handler = file.Handler(r)
	} else if strings.HasPrefix(r.URL.Path, "/icon") {
		handler = icons.Handler(r)
	} else if strings.HasPrefix(r.URL.Path, "/notification") {
		handler = notifications.Handler(r)
	} else if strings.HasPrefix(r.URL.Path, "/device") {
		handler = power.Handler(r)
	} else if strings.HasPrefix(r.URL.Path, "/search") {
		handler = search.Handler(r)
	} else if strings.HasPrefix(r.URL.Path, "/session") {
		handler = session.Handler(r)
	} else if strings.HasPrefix(r.URL.Path, "/item") {
		handler = statusnotifications.Handler(r)
	} else if strings.HasPrefix(r.URL.Path, "/watch") {
		handler = watch.Handler(r)
	} else if strings.HasPrefix(r.URL.Path, "/window") {
		handler = windows.WindowHandler(r)
	} else if r.URL.Path == "/desktoplayout" {
		handler = windows.DesktopLayoutHandler(r)
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

	go func() {
		log.Println(http.ListenAndServe("localhost:7939", nil))
	}()

	go lib.Serve("org.refude.desktop-service", http.HandlerFunc(ServeHTTP))
	_ = http.ListenAndServe(":7938", http.HandlerFunc(ServeHTTP))
}
