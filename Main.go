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
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/search"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/windows"

	"github.com/surlykke/RefudeServices/lib"

	_ "net/http/pprof"
)

var resourcefinders = []func(*http.Request) respond.JsonResource{
	applications.GetJsonResource,
	windows.GetJsonResource,
	file.GetJsonResource,
	notifications.GetJsonResource,
	statusnotifications.GetJsonResource,
	power.GetJsonResource,
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	if path == "/icon" {
		icons.ServeHTTP(w, r)
	} else if strings.HasPrefix(path, "/search/") {
		search.ServeHTTP(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/watch") {
		watch.ServeHTTP(w, r)
	} else if strings.HasPrefix(path, "/doc") {
		doc.ServeHTTP(w, r)
	} else {
		for _, resourcefinder := range resourcefinders {
			if res := resourcefinder(r); res != nil {
				ServeJsonResource(res, w, r)
				return
			}
		}
		respond.NotFound(w)
	}
}

func ServeJsonResource(jr respond.JsonResource, w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		jr.DoGet(jr, w, r)
	} else if r.Method == "POST" {
		jr.DoPost(w, r)
	} else if r.Method == "DELETE" {
		jr.DoDelete(w, r)
	} else {
		respond.NotAllowed(w)
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

	go func() {
		log.Info(http.ListenAndServe("localhost:7939", nil))
	}()

	go lib.Serve("org.refude.desktop-service", http.HandlerFunc(ServeHTTP))
	_ = http.ListenAndServe(":7938", http.HandlerFunc(ServeHTTP))
}
