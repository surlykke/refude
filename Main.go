// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package main

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/browse"
	"github.com/surlykke/RefudeServices/browsertabs"
	"github.com/surlykke/RefudeServices/config"
	"github.com/surlykke/RefudeServices/desktop"
	"github.com/surlykke/RefudeServices/doc"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/ping"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/start"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/wayland"

	_ "net/http/pprof"
)

func main() {
	log.Info("Running")

	go wayland.Run()
	go applications.Run()
	if config.Notifications.Enabled {
		go notifications.Run()
	}
	go power.Run()
	go statusnotifications.Run()

	http.Handle("/browse", browse.Handler)
	http.Handle("/browse/", browse.Handler)
	http.Handle("/ping", ping.WebsocketHandler)

	http.HandleFunc("/tabsink", browsertabs.ServeHTTP)
	http.HandleFunc("/flash", notifications.ServeFlash)
	http.HandleFunc("/file/", file.ServeHTTP)
	http.HandleFunc("/icon", icons.ServeHTTP)
	http.HandleFunc("/complete", Complete)
	http.HandleFunc("/doc", doc.ServeHTTP)
	http.HandleFunc("/start", resource.SingleResourceServer(start.Start, "/"))
	http.HandleFunc("/watch", watch.ServeHTTP)
	http.HandleFunc("/bookmarks", resource.SingleResourceServer(start.Bookmarks, "/"))
	http.HandleFunc("/desktop/", desktop.ServeHTTP)
	http.HandleFunc("/", resourcerepo.ServeHTTP)
	
	if err := http.ListenAndServe(":7938", nil); err != nil {
		log.Warn("http.ListenAndServe failed:", err)
	}
}

func Complete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, collectPaths(requests.GetSingleQueryParameter(r, "prefix", "")))
	} else {
		respond.NotAllowed(w)
	}
}

func collectPaths(prefix string) []string {
	var paths = make([]string, 0, 1000)
	paths = append(paths, "/icon?name=", "/start?search=", "/complete?prefix=", "/watch", "/doc", "/bookmarks")
	paths = append(paths, resourcerepo.GetPaths()...)

	var pos = 0
	for _, path := range paths {
		if strings.HasPrefix(path, prefix) {
			paths[pos] = path
			pos = pos + 1
		}
	}

	return paths[0:pos]
}

