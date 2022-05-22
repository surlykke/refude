// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package start

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/windows"
	"golang.org/x/exp/slices"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resource.ServeResource[string](w, r, "/start", Start{})
}

type Start struct{}

func (s Start) Id() string {
	return ""
}

func (s Start) Presentation() (title string, comment string, icon link.Href, profile string) {
	return "Start", "", "", "start"
}

func (s Start) Links(self, term string) link.List {
	return doDesktopSearch(term)
}

func doDesktopSearch(term string) link.List {
	var links = make(link.List, 0, 300)
	term = strings.ToLower(term)

	links = append(links, notifications.Search(term)...)
	links = append(links, windows.Search(term)...)

	if len(term) > 0 {
		links = append(links, file.SearchFrom(xdg.Home, term, "~/")...)
		links = append(links, applications.Search(term)...)
	}

	if len(term) > 3 {
		links = append(links, power.Search(term)...)

	}

	slices.SortFunc(links, func(l1, l2 link.Link) bool {
		if l1.Rank == l2.Rank {
			return l1.Href < l2.Href
		} else {
			return l1.Rank < l2.Rank
		}
	})

	return links
}

