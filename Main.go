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

	"github.com/surlykke/RefudeServices/lib"

	_ "net/http/pprof"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	if pathElements, ok := extractPathElements(path); !ok {
		respond.NotFound(w)
	} else {
		switch pathElements[0] {
		case "icon":
			icons.ServeHTTP(w, r)
		case "watch":
			watch.ServeHTTP(w, r)
		case "search":
			search.ServeHTTP(w, r)
		case "doc":
			doc.ServeHTTP(w, r)
		case "window":
			serveResource(w, r, windows.GetResource(pathElements[1:]))
		case "application":
			serveResource(w, r, applications.GetAppResource(pathElements[1:]))
		case "mimetype":
			serveResource(w, r, applications.GetMimeResource(pathElements[1:]))
		case "notification":
			serveResource(w, r, notifications.GetResource(pathElements[1:]))
		case "item":
			serveResource(w, r, statusnotifications.GetResource(pathElements[1:]))
		case "device":
			serveResource(w, r, power.GetResource(pathElements[1:]))
		case "file":
			serveResource(w, r, file.GetResource(pathElements[1:]))
		default:
			respond.NotFound(w)
		}
	}
}

// Checks that path is of form /foo/baa/moo (i.e starts with slash and doesn't end with slash)
// and extracts that to []string{"foo", "baa", "moo"}
func extractPathElements(path string) ([]string, bool) {
	if strings.HasPrefix(path, "/") && !strings.HasSuffix(path[1:], "/") {
		return strings.Split(path[1:], "/"), true
	} else {
		return []string{}, false
	}
}

type envelope struct {
	Links      []resource.Link `json:"_links"`
	RefudeType string          `json:"refudeType"`
	Data       interface{}     `json:"data"`
}

func serveResource(w http.ResponseWriter, r *http.Request, res resource.Resource) {
	if res == nil {
		respond.NotFound(w)
	} else {
		switch r.Method {
		case "GET":
			respond.ResourceAsJson(w, res, res.Links())
			return
		case "POST":
			if postable, ok := res.(resource.Postable); ok {
				postable.DoPost(w, r)
				return
			}
		case "DELETE":
			if deleteable, ok := res.(resource.Deleteable); ok {
				deleteable.DoDelete(w, r)
				return
			}
		}
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
