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
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
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
	http.HandleFunc("/links/", serveLinks)
	http.HandleFunc("/", serveHttp)

	if err := http.ListenAndServe(":7938", nil); err != nil {
		log.Warn("http.ListenAndServe failed:", err)
	}
}

func serveLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respond.NotAllowed(w)
	} else if res := findResource(r.URL.Path[6:]); res == nil {
		respond.NotFound(w)
	} else {
		var term = requests.GetSingleQueryParameter(r, "search", "")
		var links, filtered = res.Links(term)
		if !filtered {
			var retained = 0;
			for i := 0; i < len(links); i++ {
				if rnk := searchutils.Match(term, links[i].Title); rnk > -1 {
					links[retained] = links[i]
					links[retained].Rank = rnk
					retained++
				}
			}
			links = links[:retained]
		}
		respond.AsJson(w, links)
	}
}

func serveHttp(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	
	if path == "/start" {
		serveResource(start.Start{}, w, r)
	} else if path == "/notification/flash" {
		serveResource(notifications.GetFlashResource(), w, r)
	} else if strings.HasPrefix(path, "/file/") {
		serveResource(file.Get(path[5:]), w, r)
	} else if collection := findCollection(path); collection != nil {
		if path == collection.Prefix {
			serveList(collection.GetAll(), w, r)
		} else {
			serveResource(collection.Get(path), w, r)
		}
	} else {
		respond.NotFound(w)
	}
}

func serveList(resourceList []resource.Resource, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respond.NotAllowed(w)
	} else {
		var wrapperList = make([]jsonResource, len(resourceList), len(resourceList))
		for i := 0; i < len(resourceList); i++ {
			wrapperList[i] = makeJsonWrapper(resourceList[i])
		}
		respond.AsJson(w, wrapperList)
	}
}

func serveResource(res resource.Resource, w http.ResponseWriter, r *http.Request) {
	if res == nil {
		respond.NotFound(w)
	} else if r.Method == "GET" {
		respond.AsJson(w, makeJsonWrapper(res))
	} else if postable, ok := res.(resource.Postable); ok && r.Method == "POST" {
		postable.DoPost(w, r)
	} else if deletable, ok := res.(resource.Deleteable); ok && r.Method == "DELETE" {
		deletable.DoDelete(w, r)
	} else {
		respond.NotAllowed(w)
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

func findResource(path string) resource.Resource {
	if path == "/start" {
		return start.Start{}
	} else if strings.HasPrefix(path, "/file/") {
		return file.Get(path[5:])
	} else if collection := findCollection(path); collection != nil {
		return collection.Get(path)
	} else {
		return nil
	}
}

type jsonResource struct {
	Self    link.Href   `json:"self"`
	Links   link.Href   `json:"links,omitempty"`
	Title   string      `json:"title"`
	Comment string      `json:"comment,omitempty"`
	Icon    link.Href   `json:"icon,omitempty"`
	Profile string      `json:"profile"`
	Data    interface{} `json:"data"`
}

func makeJsonWrapper(res resource.Resource) jsonResource {
	var self = link.Href(res.Self())
	var links = link.Href("/links" + self) 
	var data = res
	var title, comment, iconName, profile = res.Presentation()
	return jsonResource{Self: self, Links: links, Title: title, Comment: comment, Icon: link.Href(iconName), Profile: profile, Data: data}
}
