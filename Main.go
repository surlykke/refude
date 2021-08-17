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

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/doc"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/search"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/windows"

	_ "net/http/pprof"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var prefix = func(pref string) bool {
		return strings.HasPrefix(r.URL.Path, pref)
	}

	switch {
	case r.URL.Path == "/icon":
		icons.ServeHTTP(w, r)
	case r.URL.Path == "/watch":
		watch.ServeHTTP(w, r)
	case prefix("/search/"):
		search.ServeHTTP(w, r)
	case prefix("/doc"):
		doc.ServeHTTP(w, r)
	case prefix("/notification/"):
		notifications.Notifications.ServeHTTP(w, r)
	case prefix("/window/"):
		windows.Windows.ServeHTTP(w, r)
	case prefix("/item/"):
		statusnotifications.Items.ServeHTTP(w, r)
	case prefix("/itemmenu/"):
		statusnotifications.Menus.ServeHTTP(w, r)
	case prefix("/device/"):
		power.Devices.ServeHTTP(w, r)
	case prefix("/application/"):
		applications.Applications.ServeHTTP(w, r)
	case prefix("/mimetype/"):
		applications.Mimetypes.ServeHTTP(w, r)
	case prefix("/file/"):
		if fileRes, ok := file.GetResource(r); ok {
			fileRes.ServeHTTP(w, r)
		} else {
			respond.NotFound(w)
		}
	default:
		respond.NotFound(w)
	}
}

func serveRes(res resource.Resource, ok bool) {
}

func main() {
	go windows.Run()
	go applications.Run()
	go notifications.Run()
	go power.Run()
	go statusnotifications.Run()
	go icons.Run()
	go watch.Run()

	if err := http.ListenAndServe(":7938", http.HandlerFunc(ServeHTTP)); err != nil {
		log.Warn("http.ListenAndServe failed:", err)
	}
}
