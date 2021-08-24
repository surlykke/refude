// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"embed"
	"net/http"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/doc"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/search"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/windows"

	_ "net/http/pprof"
)

//go:embed client
var clientResources embed.FS

var clientResourceServer = http.FileServer(http.FS(clientResources))

func main() {
	go windows.Run()
	go applications.Run()
	go notifications.Run()
	go power.Run()
	go statusnotifications.Run()
	go icons.Run()
	go watch.Run()

	http.Handle("/client/", clientResourceServer)
	http.HandleFunc("/icon", icons.ServeHTTP)
	http.HandleFunc("/search/", search.ServeHTTP)
	http.HandleFunc("/watch", watch.ServeHTTP)
	http.HandleFunc("/doc", doc.ServeHTTP)
	http.HandleFunc("/file/", file.ServeHTTP)
	http.Handle("/notification/", notifications.Notifications)
	http.Handle("/window/", windows.Windows)
	http.Handle("/item/", statusnotifications.Items)
	http.Handle("/itemmenu/", statusnotifications.Menus)
	http.Handle("/device/", power.Devices)
	http.Handle("/application/", applications.Applications)
	http.Handle("/mimetype/", applications.Mimetypes)

	if err := http.ListenAndServe(":7938", nil); err != nil {
		log.Warn("http.ListenAndServe failed:", err)
	}
}
