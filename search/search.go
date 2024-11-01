// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package search

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/surlykke/RefudeServices/desktopactions"
	"github.com/surlykke/RefudeServices/file"
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

const max_search_time = 30 * time.Millisecond

type collector func(string) []resource.Link

func Search(searchTerm string) []resource.Link {
	searchTerm = strings.ToLower(searchTerm)
	var collectors []collector

	collectors, searchTerm = collectCollectors(searchTerm)

	var result []resource.Link
	var returned = 0
	result = make([]resource.Link, 0, 100)
	var returns = make(chan []resource.Link, len(collectors))
	for i := 0; i < len(collectors); i++ {
		var j = i
		go func() {
			returns <- collectors[j](searchTerm)
		}()
	}
	var timeout = time.After(max_search_time)
	for returned < len(collectors) {
		select {
		case links := <-returns:
			result = append(result, links...)
			returned++
		case <-timeout:
			fmt.Println("Reached timeout")
			break
		}
	}

	return filterAndSort(result, searchTerm)
}

func collectCollectors(searchTerm string) ([]collector, string) {
	var fileSearchDirs = []string{xdg.Home, xdg.ConfigHome, xdg.DownloadDir, xdg.DocumentsDir, xdg.MusicDir, xdg.VideosDir}

	if strings.Index(searchTerm, "/") > -1 {
		var pathBits = strings.Split(searchTerm, "/")
		pathBits, searchTerm = pathBits[:len(pathBits)-1], pathBits[len(pathBits)-1]
		fileSearchDirs = file.CollectDirs(fileSearchDirs, pathBits) // Special handling of file Search

		return []collector{file.Collector(fileSearchDirs)}, searchTerm
	} else {
		var collectors = []collector{linkCollector("/notification/"), linkCollector("/window/"), linkCollector("/tab/")}
		if len(searchTerm) > 0 {
			collectors = append(collectors, desktopactions.GetLinks, linkCollector("/application/"))
		}
		if len(searchTerm) > 2 {
			collectors = append(collectors, linkCollector("/device/"), file.Collector(fileSearchDirs), statusnotifications.GetLinks)
		}
		return collectors, searchTerm
	}
}

func linkCollector(prefix string) collector {
	return func(string) []resource.Link {
		return repo.CollectLinks(prefix)
	}
}

func filterAndSort(links []resource.Link, term string) []resource.Link {
	if term == "" {
		return links
	} else {
		var rankedLinks = make([]resource.Link, 0, 20)
		for _, l := range links {
			l.Rank = fuzzy.RankMatchNormalizedFold(term, l.Title)
			if l.Rank < 0 {
				for _, keyWord := range l.Keywords {
					if strings.Contains(strings.ToLower(keyWord), term) {
						l.Rank = 100000
						break
					}
				}
			}
			if l.Rank > -1 {
				rankedLinks = append(rankedLinks, l)
			}
		}
		slices.SortFunc(rankedLinks, func(l1, l2 resource.Link) int {
			var tmp = l1.Rank - l2.Rank
			if tmp == 0 {
				// Not significant, just to make the sort reproducible
				tmp = strings.Compare(string(l1.Path), string(l2.Path))
			}
			return tmp
		})
		return rankedLinks
	}
}
