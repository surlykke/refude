// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package search

import (
	"slices"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/browser"
	"github.com/surlykke/RefudeServices/desktopactions"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/entity"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/response"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/wayland"
)

const maxRank uint = 1000000

func GetHandler(term string) response.Response {
	return response.Json(Search(term))
}

type searchable interface {
	OmitFromSearch() bool
	GetShort() entity.Short
}

type RankedShort struct {
	entity.Short
	Rank uint
}

func (rs *RankedShort) ActionLinks() []link.Link {
	var filtered []link.Link = make([]link.Link, 0, len(rs.Links))
	for _, l := range rs.Links {
		if l.Relation == relation.Action || l.Relation == relation.Delete {
			filtered = append(filtered, l)
		}
	}
	return filtered
}

func Search(term string) []RankedShort {
	var termRunes = []rune(strings.ToLower(term))
	var length = len(termRunes)
	var rankedShorts []RankedShort = make([]RankedShort, 0, 100)

	rankedShorts = append(rankedShorts, search(termRunes, notifications.NotificationMap.GetAll())...)
	rankedShorts = append(rankedShorts, search(termRunes, wayland.WindowMap.GetAll())...)
	rankedShorts = append(rankedShorts, search(termRunes, browser.TabMap.GetAll())...)
	if length > 0 {
		rankedShorts = append(rankedShorts, search(termRunes, applications.AppMap.GetAll())...)
		if length > 2 {
			rankedShorts = append(rankedShorts, search(termRunes, power.DeviceMap.GetAll())...)
			rankedShorts = append(rankedShorts, search(termRunes, file.FileMap.GetAll())...)
			rankedShorts = append(rankedShorts, search(termRunes, browser.BookmarkMap.GetAll())...)
			rankedShorts = append(rankedShorts, search(termRunes, desktopactions.Resources)...)
		}
	}
	slices.SortFunc(rankedShorts, func(l1, l2 RankedShort) int {
		var tmp = int(l1.Rank) - int(l2.Rank)
		if tmp == 0 {
			// Not significant, just to make the sort reproducible
			tmp = strings.Compare(string(l1.Title), string(l2.Title))
		}
		return tmp
	})

	return rankedShorts
}

func search[R searchable](term []rune, reslist []R) []RankedShort {
	var result = make([]RankedShort, 0, 100)
	for _, res := range reslist {
		if res.OmitFromSearch() {
			continue
		}

		var short = res.GetShort()
		var rank = match(short.Title, term, 0)
		for _, keyword := range short.Keywords {
			if tmp := match(keyword, term, 0); tmp < rank {
				rank = tmp
			}
		}
		if rank < maxRank {
			result = append(result, RankedShort{short, rank})
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
