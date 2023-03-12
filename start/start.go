// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package start

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

type Start struct{}

func Get(path string) resource.Resource {
	return Start{}
}

func (s Start) Path() string {
	return ""
}

func (s Start) Presentation() (title string, comment string, icon link.Href, profile string) {
	return "Start", "", "", "start"
}

func (s Start) Links(self, term string) link.List {
	return doDesktopSearch(term)
}

func (s Start) RelevantForSearch() bool {
	return true
}

func doDesktopSearch(term string) link.List {
	var links = make(link.List, 0, 300)
	term = strings.ToLower(term)

	// Could be done concurrently..
	links = append(links, rewriteAndSort("/notification/", notifications.Notifications.Search(term, 0))...)
	links = append(links, rewriteAndSort("/window/", windows.GetResourceRepo().Search(term, 0))...)
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
	rewritten.SortByRank()
	return rewritten
}
