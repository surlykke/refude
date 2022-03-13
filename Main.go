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
	"github.com/surlykke/RefudeServices/lib/requests"
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
	go windows.Run()
	go applications.Run()
	go notifications.Run()
	go power.Run()
	go statusnotifications.Run()

	http.HandleFunc("/refude/", client.ServeHTTP)
	http.HandleFunc("/icon", icons.ServeHTTP)
	http.HandleFunc("/complete", complete.ServeHTTP)
	http.HandleFunc("/watch", watch.ServeHTTP)
	http.HandleFunc("/doc", doc.ServeHTTP)
	http.HandleFunc("/", serveHttp)

	if err := http.ListenAndServe(":7938", nil); err != nil {
		log.Warn("http.ListenAndServe failed:", err)
	}
}

func serveHttp(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path

	if path == "/start" {
		serveResource(w, r, start.Start{})
	} else if path == "/notification/flash" {
		serveResource(w, r, notifications.GetFlashResource())
	} else if strings.HasPrefix(path, "/file/") {
		serveResource(w, r, file.Get(path[5:]))
	} else if collection := findCollection(path); collection != nil {
		if path == collection.Prefix {
			serveList(w, r, collection.GetAll())
		} else {
			serveResource(w, r, collection.Get(path))
		}
	} else {
		respond.NotFound(w)
	}

}

func serveList(w http.ResponseWriter, r *http.Request, resources []resource.Resource) {
	if r.Method != "GET" {
		respond.NotAllowed(w)
	} else {
		var wrapperList = make([]resource.Wrapper, len(resources), len(resources))
		for i, res := range resources {
			wrapperList[i] = resource.MakeWrapper(res, "")
		}
		respond.AsJson(w, wrapperList)
	}
}

func serveResource(w http.ResponseWriter, r *http.Request, res resource.Resource) {
	if res == nil {
		respond.NotFound(w)
	} else {
		var linkSearchTerm = requests.GetSingleQueryParameter(r, "search", "")
		var wrapper = resource.MakeWrapper(res, linkSearchTerm)
		if r.Method == "GET" {
			respond.AsJson(w, wrapper)
		} else if postable, ok := res.(resource.Postable); ok && r.Method == "POST" {
			postable.DoPost(w, r)
		} else if deletable, ok := res.(resource.Deleteable); ok && r.Method == "DELETE" {
			deletable.DoDelete(w, r)
		} else {
			respond.NotAllowed(w)
		}
	}
}

func findCollection(path string) *resource.Collection {
	if strings.HasPrefix(path, "/notification/") {
		return notifications.Notifications
	} else if strings.HasPrefix(path, "/icontheme/") {
		return icons.IconThemes
	} else if strings.HasPrefix(path, "/window/") {
		return windows.Windows
	} else if strings.HasPrefix(path, "/item/") {
		return statusnotifications.Items
	} else if strings.HasPrefix(path, "/itemmenu/") {
		return statusnotifications.Menus
	} else if strings.HasPrefix(path, "/device/") {
		return power.Devices
	} else if strings.HasPrefix(path, "/application/") {
		return applications.Applications
	} else if strings.HasPrefix(path, "/mimetype/") {
		return applications.Mimetypes
	} else {
		return nil
	}
}
