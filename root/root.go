// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package root

import (
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/windows"
)

type RootRepo struct{}

func (this RootRepo) GetResources() []resource.Resource {
	return []resource.Resource{&start, &bookmarks}
}

func (this RootRepo) GetResource(path string) resource.Resource {
	switch path {
	case "start":
		return &start
	case "bookmarks":
		return &bookmarks
	default:
		return nil
	}
}

func (this RootRepo) Search(term string, threshold int) link.List {
	return link.List{}
}

var Repo RootRepo

type Start struct {
	resource.BaseResource
	searchTerm string
}

func (s *Start) Links(term string) link.List {
	return doDesktopSearch(term)
}

func doDesktopSearch(term string) link.List {
	var links = make(link.List, 0, 300)
	term = strings.ToLower(term)

	// Could perhaps be done concurrently..
	links = append(links, rewriteAndSort("/notification/", notifications.Notifications.Search(term, 0))...)
	links = append(links, rewriteAndSort("/window/", windows.GetWindowCollection().Search(term, 0))...)
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

var start = Start{
	BaseResource: resource.BaseResource{Path: "start", Title: "Start", Profile: "start"},
}


type Bookmarks struct {
	resource.BaseResource
}


func (bm Bookmarks) Links(searchTerm string) link.List {
   return link.List{
	{ Href: "/application/", Title: "Applications", Profile: "application*"},
	{ Href: "/window/", Title: "Windows", Profile: "window*"},
	{ Href: "/notification/", Title: "Notifications", Profile: "notification*"},
	{ Href: "/device/", Title: "Devices", Profile: "device*"},
	{ Href: "/item/", Title: "Items", Profile: "item*"}}
}

var bookmarks = Bookmarks{BaseResource:resource.BaseResource{Path: "bookmarks", Title: "Bookmarks", Profile: "bookmarks"}}
