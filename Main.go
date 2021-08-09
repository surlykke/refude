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

	var pathAfter = func(prefix string) (string, bool) {
		if strings.HasPrefix(path, prefix) {
			return path[len(prefix):], true
		} else {
			return "", false
		}
	}

	var s string
	var ok bool

	if path == "/icon" {
		icons.ServeHTTP(w, r)
	} else if path == "/watch" {
		watch.ServeHTTP(w, r)
	} else if _, ok = pathAfter("/search/"); ok {
		search.ServeHTTP(w, r)
	} else if _, ok = pathAfter("/doc"); ok {
		doc.ServeHTTP(w, r)
	} else if s, ok = pathAfter("/window/"); ok {
		serveResource(w, r, windows.GetResource(s))
	} else if s, ok = pathAfter("/application/"); ok {
		serveResource(w, r, applications.GetAppResource(s))
	} else if s, ok = pathAfter("/mimetype/"); ok {
		serveResource(w, r, applications.GetMimeResource(s))
	} else if s, ok = pathAfter("/notification/"); ok {
		serveResource(w, r, notifications.GetResource(s))
	} else if s, ok = pathAfter("/item/"); ok {
		serveResource(w, r, statusnotifications.GetResource(s))
	} else if s, ok = pathAfter("/device/"); ok {
		serveResource(w, r, power.GetResource(s))
	} else if s, ok = pathAfter("/file/"); ok {
		serveResource(w, r, file.GetResource(s))
	} else {
		respond.NotFound(w)
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

	if err := http.ListenAndServe(":7938", http.HandlerFunc(ServeHTTP)); err != nil {
		log.Warn("http.ListenAndServe failed:", err)
	}
}
