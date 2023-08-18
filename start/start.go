// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package start

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/pubsub"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/wayland"
	"github.com/surlykke/RefudeServices/x11"
)

var searchChanges = pubsub.MakePublisher[string]()

func watch(subscription *pubsub.Subscription[string]) {
	for {
		searchChanges.Publish(subscription.Next())
	}
}

var Run = func() {
	go watch(notifications.Notifications.Subscribe())
	go watch(x11.Windows.Subscribe())
	go watch(applications.Applications.Subscribe())
	go watch(power.Devices.Subscribe())
}

type StartResource struct {
	resource.BaseResource
	searchTerm string
}

func (s *StartResource) Links(term string) link.List {
	return doDesktopSearch(term)
}

func doDesktopSearch(term string) link.List {
	var links = make(link.List, 0, 300)
	term = strings.ToLower(term)

	// Could perhaps be done concurrently..
	links = append(links, rewriteAndSort("/notification/", notifications.Notifications.Search(term, 0))...)
	if xdg.SessionType == "x11" {
		links = append(links, rewriteAndSort("/window/", x11.Windows.Search(term, 0))...)
	} else {
		links = append(links, rewriteAndSort("/window/", wayland.Windows.Search(term, 0))...)
	}
	links = append(links, rewriteAndSort("/application/", applications.Applications.Search(term, 1))...)
	links = append(links, rewriteAndSort("/file/", file.FileRepo.Search(term, 2))...)
	links = append(links, rewriteAndSort("/device/", power.Devices.Search(term, 3))...)

	return links
}

func rewriteAndSort(context string, links link.List) link.List {
	var rewritten = make(link.List, 0, len(links))
	for _, lnk := range links {
		var tmp = lnk
		tmp.Href = link.Href(context) + tmp.Href
		rewritten = append(rewritten, tmp)
	}
	return rewritten
}

var Start = &StartResource{
	BaseResource: resource.BaseResource{Id: "start", Title: "Start", Profile: "start"},
}

type BookmarksResource struct {
	resource.BaseResource
}

func (bm BookmarksResource) Links(searchTerm string) link.List {
	return link.List{
		{Href: "/application/", Title: "Applications", Profile: "application*"},
		{Href: "/window/", Title: "Windows", Profile: "window*"},
		{Href: "/notification/", Title: "Notifications", Profile: "notification*"},
		{Href: "/device/", Title: "Devices", Profile: "device*"},
		{Href: "/item/", Title: "Items", Profile: "item*"}}
}

var Bookmarks = &BookmarksResource{BaseResource: resource.BaseResource{Id: "bookmarks", Title: "Bookmarks", Profile: "bookmarks"}}


func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respond.NotAllowed(w)
	} else if r.URL.Path == "/start/watch" {
		var subscription = searchChanges.Subscribe()

		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.(http.Flusher).Flush()

		if _, err := fmt.Fprintf(w, "data:%s\n\n", ""); err != nil {
			return
		}
		w.(http.Flusher).Flush()

		for {
			if _, err := fmt.Fprintf(w, "data:%s\n\n", subscription.Next()); err != nil {
				return
			}
			w.(http.Flusher).Flush()
		}

	} else {
		respond.NotFound(w)
	}
}
