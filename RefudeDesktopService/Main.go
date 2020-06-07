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

// Associate handler with path prefix
var handlers = map[string]func(http.ResponseWriter, *http.Request){
	"/window":       windows.ServeHTTP,
	"/application":  applications.ServeHTTP,
	"/mimetype":     applications.ServeHTTP,
	"/notification": notifications.ServeHTTP,
	"/device":       power.ServeHTTP,
	"/item":         statusnotifications.ServeHTTP,
	"/icon":         icons.ServeHTTP,
	"/session":      session.ServeHTTP,
	"/search":       search.ServeHTTP,
	"/file":         file.ServeHTTP,
	"/watch":        watch.ServeHTTP,
	"/doc":          doc.ServeHTTP,
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for prefix, handler := range handlers {
		if strings.HasPrefix(r.URL.Path, prefix) {
			handler(w, r)
			return
		}
	}
	respond.NotFound(w)
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
