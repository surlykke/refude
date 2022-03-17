// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/client"
	"github.com/surlykke/RefudeServices/complete"
	"github.com/surlykke/RefudeServices/doc"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/start"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/windows"

	_ "net/http/pprof"
)

func FallBack(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Fallback:", r.Method, r.URL.Path)
	respond.NotFound(w)
}

func main() {
	go applications.Run()
	go notifications.Run()
	go power.Run()
	go statusnotifications.Run()

	http.HandleFunc("/refude/", client.ServeHTTP)
	http.HandleFunc("/icon", icons.ServeHTTP)
	http.HandleFunc("/complete", complete.ServeHTTP)
	http.HandleFunc("/watch", watch.ServeHTTP)
	http.HandleFunc("/doc", doc.ServeHTTP)
	http.HandleFunc("/window/", windows.ServeHTTP)
	http.Handle("/notification/", notifications.Notifications)
	http.Handle("/icontheme/", icons.IconThemes)
	http.Handle("/item/", statusnotifications.Items)
	http.Handle("/itemmenu/", statusnotifications.Menus)
	http.Handle("/device/", power.Devices)
	http.Handle("/application/", applications.Applications)
	http.Handle("/mimetype/", applications.Mimetypes)
	http.HandleFunc("/", serveHttp)

	if err := http.ListenAndServe(":7938", nil); err != nil {
		log.Warn("http.ListenAndServe failed:", err)
	}
}

func serveHttp(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path

	if path == "/start" {
		resource.ServeResource(w, r, start.Start{})
	} else if strings.HasPrefix(path, "/file/") {
		resource.ServeResource(w, r, file.Get(path[5:]))
	} else {
		respond.NotFound(w)
	}

}

