// Copyright (c) 2017 Christian Surlykke
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
	"github.com/surlykke/RefudeServices/doc"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
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
	var path = r.URL.EscapedPath()
	// We only serve paths in the form '/foo/baa/moo', ie. starts with a slash,
	// then a number of pathelements (eg. foo,baa,moo) separated by slash, and no ending slash
	if !strings.HasPrefix(path, "/") || strings.HasSuffix(path, "/") {
		respond.NotFound(w)
	} else {
		var pathElements = strings.Split(path[1:], "/")
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
			fmt.Println("Looking for file")
			serveResource(w, r, file.GetResource(pathElements[1:]))
		default:
			respond.NotFound(w)
		}
	}
}

func serveResource(w http.ResponseWriter, r *http.Request, res resource.Resource) {
	if res == nil {
		respond.NotFound(w)
	} else {
		switch r.Method {
		case "GET":
			var links = search.Filter(res.Links(), requests.Term(r))
			respond.ResourceAsJson(w, links, res.RefudeType(), res)
			return
		case "POST":
			if postable, ok := res.(resource.Postable); ok {
				fmt.Println("postable...")
				postable.DoPost(w, r)
				return
			} else {
				fmt.Println("not postable...")
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

	if err := http.ListenAndServe(":7938", http.HandlerFunc(ServeHTTP)); err != nil {
		log.Warn("http.ListenAndServe failed:", err)
	}
}
