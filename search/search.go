// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package search

import (
	"net/http"
	"strings"

	"github.com/sahilm/fuzzy"
	"github.com/surlykke/RefudeServices/desktopactions"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/statusnotifications"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var term = strings.ToLower(requests.GetSingleQueryParameter(r, "term", ""))
		respond.AsJson(w, Search(term))
	} else {
		respond.NotAllowed(w)
	}
}

func Search(term string) []resource.Link {
	var result = make([]resource.Link, 0, 100)
	var fileDirs = []string{xdg.Home, xdg.ConfigHome, xdg.DownloadDir, xdg.DocumentsDir, xdg.MusicDir, xdg.VideosDir}
	if strings.Index(term, "/") > -1 {
		var pathBits = strings.Split(term, "/")
		pathBits, term = pathBits[:len(pathBits)-1], pathBits[len(pathBits)-1]
		fileDirs = file.CollectDirs(fileDirs, pathBits)
		for _, dir := range fileDirs {
			file.Collect(&result, dir)
		}
	} else {

		getLinks(&result, "/notification/")
		getLinks(&result, "/window/")
		getLinks(&result, "/tab/")

		if len(term) > 0 {
			getStartLinks(&result)
			getLinks(&result, "/application/")
		}

		if len(term) > 2 {
			getLinks(&result, "/device/")
			result = append(result, file.MakeLinkFromPath(xdg.Home, "Home"))
			for _, dir := range fileDirs {
				file.Collect(&result, dir)
			}
			getMenuLinks(&result)
		}
	}

	return filterAndSort(result, term)
}

func getLinks(collector *[]resource.Link, prefix string) {
	for _, res := range repo.GetListUntyped(prefix) {
		if !res.OmitFromSearch() {
			*collector = append(*collector, resource.LinkTo(res))
		}
	}
}

func getStartLinks(collector *[]resource.Link) {
	if start, ok := repo.Get[*desktopactions.StartResource]("/start"); ok {
		*collector = append(*collector, start.GetLinks(relation.Action)...)
	}
}

func getMenuLinks(collector *[]resource.Link) {
	for _, itemMenu := range repo.GetList[*statusnotifications.Menu]("/menu/") {
		if entries, err := itemMenu.Entries(); err == nil {
			getLinksFromMenu(collector, itemMenu.Path, entries)
		}
	}
}

func getLinksFromMenu(collector *[]resource.Link, menuPath string, entries []statusnotifications.MenuEntry) {
	for _, entry := range entries {
		if entry.Type == "standard" {
			if len(entry.SubEntries) > 0 {
				getLinksFromMenu(collector, menuPath, entry.SubEntries)
			} else {
				var href = menuPath + "?id=" + entry.Id
				*collector = append(*collector, resource.Link{Href: href, Title: entry.Label, IconUrl: entry.IconUrl, Relation: relation.Action})
			}
		}
	}
}

type linkList []resource.Link

// Implement fuzzy.Source

func (ll linkList) String(i int) string {
	return ll[i].Title
}

func (ll linkList) Len() int {
	return len(ll)
}

func filterAndSort(links []resource.Link, term string) []resource.Link {
	if term == "" {
		return links
	} else {
		var ll = linkList(links)
		if lastSlash := strings.LastIndex(term, "/"); lastSlash > -1 {
			term = term[lastSlash+1:]
		}
		var matches = fuzzy.FindFrom(term, ll)
		var sorted = make([]resource.Link, len(matches), len(matches))
		for i, match := range matches {
			sorted[i] = ll[match.Index]
		}
		return sorted
	}
}
