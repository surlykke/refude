// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package main

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/browse"
	"github.com/surlykke/RefudeServices/client"
	"github.com/surlykke/RefudeServices/config"
	"github.com/surlykke/RefudeServices/doc"
	"github.com/surlykke/RefudeServices/root"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/windows"
	"github.com/surlykke/RefudeServices/windows/monitor"

	_ "net/http/pprof"
)

func main() {
	log.Info("Running")

	go windows.Run()
	go applications.Run()
	if config.Notifications.Enabled {
		go notifications.Run()
	}
	go power.Run()
	//go statusnotifications.Run()

	http.HandleFunc("/window/", collectionHandler("/window/", windows.GetResourceRepo()))
	http.HandleFunc("/application/", collectionHandler("/application/", applications.Applications))
	http.HandleFunc("/notification/", collectionHandler("/notification/", notifications.Notifications))
	http.HandleFunc("/notification/flash", resourceHandler("/notification/", notifications.GetFlash))
	http.HandleFunc("/device/", collectionHandler("/device/", power.Devices))
	http.HandleFunc("/file/", collectionHandler("/file/", file.FileRepo))
	http.HandleFunc("/icontheme/", collectionHandler("/icontheme/", icons.IconThemes))
	http.HandleFunc("/item/", collectionHandler("/item/", statusnotifications.Items))
	http.HandleFunc("/itemmenu/", collectionHandler("/itemmenu/", statusnotifications.Menus))
	http.HandleFunc("/mimetype/", collectionHandler("/mimetype/", applications.Mimetypes))
	http.HandleFunc("/screen/", collectionHandler("/screen/", monitor.Repo))
	http.HandleFunc("/", collectionHandler("/", root.Repo))

	http.HandleFunc("/refude/", client.ServeHTTP)
	http.HandleFunc("/icon", icons.ServeHTTP)
	http.HandleFunc("/complete", Complete)
	http.HandleFunc("/watch", watch.ServeHTTP)
	http.HandleFunc("/doc", doc.ServeHTTP)

	http.Handle("/browse", browse.Handler)
	http.Handle("/browse/", browse.Handler)

	if err := http.ListenAndServe(":7938", nil); err != nil {
		log.Warn("http.ListenAndServe failed:", err)
	}
}

func collectionHandler(context string, resourceRepo resource.ResourceRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serve(w, r, resourceRepo, context)
	}
}

func resourceHandler(context string, resourceGetter func(string) resource.Resource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serveSingle(w, r, resourceGetter, context)
	}
}

func handle(context string, resourceRepo resource.ResourceRepo) {
	http.HandleFunc(context, func(w http.ResponseWriter, r *http.Request) {
		serve(w, r, resourceRepo, context)
	})
}

func handleSingle(path string, getter func(path string) resource.Resource) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		serveSingle(w, r, getter, "")
	})
}

func serve(w http.ResponseWriter, r *http.Request, resourceRepo resource.ResourceRepo, context string) {
	if r.URL.Path == context {
		// Serve list
		var resources = resourceRepo.GetResources()
		if r.Method == "GET" {
			var wrappers = make([]wrapper, 0, len(resources))
			for _, res := range resources {
				wrappers = append(wrappers, makeWrapper(res, context, ""))
			}
			respond.AsJson(w, wrappers)
		} else {
			respond.NotAllowed(w)
		}
	} else {
		serveSingle(w, r, resourceRepo.GetResource, context)
	}

}

// Caller ensures r.URL.Path starts with context
func serveSingle(w http.ResponseWriter, r *http.Request, getter func(string) resource.Resource, context string) {
	var path = r.URL.Path[len(context):]
	var res = getter(path)
	if res == nil {
		respond.NotFound(w)
	} else if r.Method == "GET" {
		var linkSearchTerm = requests.GetSingleQueryParameter(r, "search", "")
		respond.AsJson(w, makeWrapper(res, context, linkSearchTerm))
	} else if postable, ok := res.(resource.Postable); ok && r.Method == "POST" {
		postable.DoPost(w, r)
	} else if deletable, ok := res.(resource.Deleteable); ok && r.Method == "DELETE" {
		deletable.DoDelete(w, r)
	} else {
		respond.NotAllowed(w)
	}
}

type wrapper struct {
	Self    link.Href   `json:"self"`
	Links   link.List   `json:"links"`
	Title   string      `json:"title"`
	Comment string      `json:"comment,omitempty"`
	Icon    link.Href   `json:"icon,omitempty"`
	Profile string      `json:"profile"`
	Data    interface{} `json:"data"`
}

func makeWrapper(res resource.Resource, context, searchTerm string) wrapper {
	var wrapper = wrapper{}
	wrapper.Self = link.Href(context + res.GetPath())
	wrapper.Links = buildFilterAndRewriteLinks(res, context, searchTerm) 
	wrapper.Data = res
	var iconName string
	wrapper.Title, wrapper.Comment, iconName, wrapper.Profile = res.Presentation()
	wrapper.Icon = link.IconUrl(iconName)
	return wrapper
}

func buildFilterAndRewriteLinks(res resource.Resource, context, searchTerm string) link.List {
	var list = make(link.List, 0, 10)
	for _, action := range res.Actions() {
		var href = context + res.GetPath()
		if action.Name != "" {
			href += "?action=" + action.Name
		}
		if searchutils.Match(searchTerm, action.Name) < 0 {
			continue 
		}
		list = append(list, link.Make(href, action.Title, action.IconName, relation.Action))
	}
	if deleteTitle, ok := res.DeleteAction(); ok {
		if searchutils.Match(searchTerm, deleteTitle) > -1 {
			list = append(list, link.Make(context + res.GetPath(), deleteTitle, "", relation.Delete))
		}
	}

	var lnks link.List = res.Links(searchTerm)

	for _, lnk := range lnks {
		if ! strings.HasPrefix(string(lnk.Href), "/") {
			lnk.Href = link.Href(context) + lnk.Href
		} 
		list = append(list, lnk)
	}

	return list
}


func Complete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, collectPaths(requests.GetSingleQueryParameter(r, "prefix", "")))
	} else {
		respond.NotAllowed(w)
	}
}

func collectPaths(prefix string) []string {
	var paths = make([]string, 0, 1000)
	paths = append(paths, "/icon?name=", "/start?search=", "/complete?prefix=", "/watch", "/doc", "/bookmarks")
	paths = append(paths, rewrite("/window/", windows.GetPaths())...)
	paths = append(paths, rewrite("/application/", applications.Applications.GetPaths())...)
	paths = append(paths, rewrite("/mimetype/", applications.Mimetypes.GetPaths())...)
	paths = append(paths, rewrite("/item/", statusnotifications.Items.GetPaths())...)
	paths = append(paths, rewrite("/notification/", notifications.Notifications.GetPaths())...)
	paths = append(paths, rewrite("/device/", power.Devices.GetPaths())...)
	paths = append(paths, rewrite("/icontheme/", icons.IconThemes.GetPaths())...)

	var pos = 0
	for _, path := range paths {
		if strings.HasPrefix(path, prefix) {
			paths[pos] = path
			pos = pos + 1
		}
	}

	return paths[0:pos]
}

func rewrite(context string, paths []string) []string {
	var rewritten = make([]string, 0, len(paths))
	for _, path := range paths {
		rewritten = append(rewritten, context+path)
	}
	return rewritten
}
