// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package search

import (
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/statusnotifications"
)

const maxRank uint = 1000000

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var term = requests.GetSingleQueryParameter(r, "term", "")
		respond.AsJson(w, Search(term))
	} else {
		respond.NotAllowed(w)
	}
}

const max_search_time = 30 * time.Millisecond

type rankedLink struct {
	resource.Link
	rank uint
}

func Search(term string) []rankedLink {
	var termRunes = []rune(strings.ToLower(term))
	var links []rankedLink
	if len(term) == 0 {
		links = searchResources(repo.GetListUntyped("/notification/", "/window/", "/tab/"), termRunes)
	} else if len(term) == 1 {
		links = searchResources(repo.GetListUntyped("/notification/", "/window/", "/tab/", "/start", "/application"), termRunes)
	} else {
		links = searchResources(repo.GetListUntyped("/notification/", "/window/", "/tab/", "/start", "/application", "/device"), termRunes)
		links = append(links, searchLinks(statusnotifications.GetLinks(), termRunes)...)
		links = append(links, searchResources(file.GetFiles(), termRunes)...)
	}

	slices.SortFunc(links, func(l1, l2 rankedLink) int {
		var tmp = int(l1.rank) - int(l2.rank)
		if tmp == 0 {
			// Not significant, just to make the sort reproducible
			tmp = strings.Compare(string(l1.Href), string(l2.Href))
		}
		return tmp
	})

	return links
}

func searchResources(reslist []resource.Resource, term []rune) []rankedLink {
	var result = make([]rankedLink, 0, 2*len(reslist))
	for _, res := range reslist {
		if res.OmitFromSearch() {
			continue
		}
		var resRank = match(res.Data().Link().Title, term, 100)
		for _, keyword := range res.Data().Keywords {
			if tmp := match(keyword, term, 100); tmp < resRank {
				resRank = tmp
			}
		}

		var correction uint = 0
		for _, l := range res.Data().GetLinks(relation.Action, relation.Delete) {
			if linkRank := match(l.Title, term, correction); linkRank < maxRank {
				result = append(result, rankedLink{Link: l, rank: linkRank})
			} else if resRank < maxRank {
				result = append(result, rankedLink{Link: l, rank: resRank})
			}
			correction = 20
		}
	}
	return result
}

func searchLinks(linkList []resource.Link, term []rune) []rankedLink {
	var result = make([]rankedLink, 0, len(linkList))
	for _, l := range linkList {
		if rnk := match(l.Title, term, 0); rnk < maxRank {
			result = append(result, rankedLink{Link: l, rank: rnk})
		}
	}
	return result
}

// Kindof 'has Substring with skips'. So eg. 'nvim' matches 'neovim' or 'pwr' matches 'poweroff'
func match(text string, term []rune, correction uint) uint {
	if len(term) == 0 {
		return 0
	}

	var rnk uint = 0

	text = strings.ToLower(text)
	var j, lastPos = 0, -1
	for i, r := range text {
		if term[j] == r {
			if j == 0 {
				rnk += 5 * uint(i-lastPos-1)
			} else {
				rnk += uint(i - lastPos - 1)
			}
			lastPos = i
			j++
		}
		if j >= len(term) {
			return rnk + correction
		}
	}

	return maxRank
}
