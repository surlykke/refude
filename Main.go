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
	"github.com/surlykke/RefudeServices/client"
	"github.com/surlykke/RefudeServices/config"
	"github.com/surlykke/RefudeServices/doc"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/ping"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/start"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/wayland"
	"github.com/surlykke/RefudeServices/x11"

	_ "net/http/pprof"
)

func main() {
	log.Info("Running")

	if xdg.SessionType == "x11" {
		go x11.Run()
	} else {
		go wayland.Run()
	}
	go applications.Run()
	if config.Notifications.Enabled {
		go notifications.Run()
	}
	go power.Run()
	//go statusnotifications.Run()
	go start.Run()

	if xdg.SessionType == "x11" {
		http.Handle("/window/", x11.Windows)
	} else {
		http.Handle("/window/", wayland.Windows)
	}
	http.Handle("/application/", applications.Applications)
	http.Handle("/notification/", notifications.Notifications)
	http.Handle("/notification/websocket", notifications.WebsocketHandler)
	http.Handle("/device/", power.Devices)
	http.Handle("/icontheme/", icons.IconThemes)
	http.Handle("/item/", statusnotifications.Items)
	http.Handle("/itemmenu/", statusnotifications.Menus)
	http.Handle("/mimetype/", applications.Mimetypes)
	http.Handle("/browse", browse.Handler)
	http.Handle("/browse/", browse.Handler)
	http.Handle("/refude/", client.StaticServer)
	http.Handle("/tab/", browsertabs.Tabs)
	http.Handle("/tab/websocket", browsertabs.WebsocketHandler)
	http.Handle("/ping", ping.WebsocketHandler)
	http.HandleFunc("/file/", file.ServeHTTP)
	http.HandleFunc("/icon", icons.ServeHTTP)
	http.HandleFunc("/complete", Complete)
	http.HandleFunc("/doc", doc.ServeHTTP)
	http.HandleFunc("/start", resource.SingleResourceServer(start.Start, "/"))
	http.HandleFunc("/start/watch", start.ServeHTTP)
	http.HandleFunc("/bookmarks", resource.SingleResourceServer(start.Bookmarks, "/"))
	
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
	if xdg.SessionType == "x11" {
		paths = append(paths, x11.Windows.GetPaths()...)
	} else {
		paths = append(paths, wayland.Windows.GetPaths()...)
	}
	paths = append(paths, applications.Applications.GetPaths()...)
	paths = append(paths, applications.Mimetypes.GetPaths()...)
	paths = append(paths, statusnotifications.Items.GetPaths()...)
	paths = append(paths, notifications.Notifications.GetPaths()...)
	paths = append(paths, power.Devices.GetPaths()...)
	paths = append(paths, icons.IconThemes.GetPaths()...)

	var pos = 0
	for _, path := range paths {
		if strings.HasPrefix(path, prefix) {
			paths[pos] = path
			pos = pos + 1
		}
	}

	return paths[0:pos]
}

