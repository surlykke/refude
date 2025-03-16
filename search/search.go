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
	"github.com/surlykke/RefudeServices/lib/response"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/wayland"
)

const maxRank uint = 1000000

func GetHandler(term string) response.Response {
	return response.Json(Search(term))
}

type Ranked struct {
	entity.Base
	Rank uint
}

func Search(term string) []Ranked {
	var termRunes = []rune(strings.ToLower(term))
	var length = len(termRunes)
	var bases = make([]entity.Base, 0, 1000)

	notifications.NotificationMap.GetForSearch(&bases)
	wayland.WindowMap.GetForSearch(&bases)
	browser.TabMap.GetForSearch(&bases)

	if length > 0 {
		applications.AppMap.GetForSearch(&bases)
		if length > 2 {
			power.DeviceMap.GetForSearch(&bases)
			file.FileMap.GetForSearch(&bases)
			browser.BookmarkMap.GetForSearch(&bases)
			bases = append(bases, *desktopactions.Start.GetBase())
		}
	}

	var result = make([]Ranked, 0, len(bases))
	for _, res := range bases {
		var rank = match(res.Title, termRunes, 0)
		for _, keyword := range res.Keywords {
			if tmp := match(keyword, termRunes, 0); tmp < rank {
				rank = tmp
			}
		}
		if rank < maxRank {
			result = append(result, Ranked{res, rank})
		}

	}

	slices.SortFunc(result, func(l1, l2 Ranked) int {
		var tmp = int(l1.Rank) - int(l2.Rank)
		if tmp == 0 {
			// Not significant, just to make the sort reproducible
			tmp = strings.Compare(string(l1.Title), string(l2.Title))
		}
		return tmp
	})

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
