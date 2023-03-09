// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package main

import (
	"net/http"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/bookmarks"
	"github.com/surlykke/RefudeServices/browse"
	"github.com/surlykke/RefudeServices/client"
	"github.com/surlykke/RefudeServices/complete"
	"github.com/surlykke/RefudeServices/config"
	"github.com/surlykke/RefudeServices/doc"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/start"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/windows"

	_ "net/http/pprof"
)

func main() {
	log.Info("Running")

	go windows.WM.Run()
	go applications.Run()
	if config.Notifications.Enabled {
		go notifications.Run()
	}
	go power.Run()
	//go statusnotifications.Run()

	http.HandleFunc("/start", start.ServeHTTP)
	http.HandleFunc("/refude/", client.ServeHTTP)
	http.HandleFunc("/icon", icons.ServeHTTP)
	http.HandleFunc("/icontheme/", icons.ServeHTTP)
	http.HandleFunc("/iconthemes", icons.ServeHTTP)
	http.HandleFunc("/complete", complete.ServeHTTP)
	http.HandleFunc("/watch", watch.ServeHTTP)
	http.HandleFunc("/doc", doc.ServeHTTP)
	http.HandleFunc("/file/", file.ServeHTTP)
	http.HandleFunc("/tmux/", windows.ServeHTTP)
	http.HandleFunc("/bookmarks", bookmarks.ServeHTTP)
	http.HandleFunc("/bookmarks/", bookmarks.ServeHTTP)
	if config.Notifications.Enabled {
		http.HandleFunc("/notification/", notifications.ServeHTTP)
	}
	http.Handle("/browse", browse.Handler)
	http.Handle("/browse/", browse.Handler)
	http.Handle("/window/", windows.WM)
	http.Handle("/item/", statusnotifications.Items)
	http.Handle("/itemmenu/", statusnotifications.Menus)
	http.Handle("/device/", power.Devices)
	http.Handle("/application/", applications.Applications)
	http.Handle("/mimetype/", applications.Mimetypes)

	if err := http.ListenAndServe(":7938", nil); err != nil {
		log.Warn("http.ListenAndServe failed:", err)
	}
}
