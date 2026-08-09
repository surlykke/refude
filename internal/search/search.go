// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package search

import (
	"net/http"
	"slices"
	"strings"

	"github.com/surlykke/refude/internal/lib/resource"
	"github.com/surlykke/refude/internal/lib/respond"
	"github.com/surlykke/refude/internal/lib/utils"
)

func Run() {
	http.HandleFunc("GET /search", func(w http.ResponseWriter, r *http.Request) {
		respond.AsJson(w, Search(utils.QueryParam(r, "term")))
	})
}

const maxRank uint = 1000000

type Ranked struct {
	resource.Base
	Rank uint `json:"-"`
}

func Search(term string) []Ranked {
	var m = makeMatcher(term)
	var result = make([]Ranked, 0, 1000)

	var prefixes = []string{"/window/", "/tab/", "/notification/"}
	if len(m.term) > 0 {
		prefixes = append(prefixes, "/application/")
	}
	if len(m.term) > 1 {
		prefixes = append(prefixes, "/expirednotification/")
	}
	if len(m.term) > 2 {
		prefixes = append(prefixes, "/network/", "/device/", "/file/", "/start/") // TODO: /bookmark/
	}
	result = filter(resource.GetForSearch(prefixes...), m)
	sort(result)
	return result
}

func filter(bases []resource.Base, m matcher) []Ranked {
	var result = make([]Ranked, 0, len(bases))
	for _, res := range bases {
		var rankCalculated = m.match(res.Title)
		for _, keyword := range res.Keywords {
			if tmp := m.match(keyword) + 20; tmp < rankCalculated {
				rankCalculated = tmp
			}
		}
		if rankCalculated < maxRank {
			result = append(result, Ranked{Base: res, Rank: rankCalculated})
		}

	}
	return result
}

func sort(list []Ranked) {
	slices.SortFunc(list, func(l1, l2 Ranked) int {
		var tmp = int(l1.Rank) - int(l2.Rank)
		if tmp == 0 {
			// Not significant, just to make the sort reproducible
			tmp = strings.Compare(string(l1.Title), string(l2.Title))
		}
		return tmp
	})

}
