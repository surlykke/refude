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
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/entity"
	"github.com/surlykke/RefudeServices/lib/mediatype"
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
	entity.Base // Basen dether
	Rank        uint
}

func SearchType(termRunes []rune, mtype mediatype.MediaType) []Ranked {
	switch mtype {
	case mediatype.Notification:
		return filter(termRunes, notifications.NotificationMap.GetForSearch())
	case mediatype.Window:
		return filter(termRunes, wayland.WindowMap.GetForSearch())
	case mediatype.Tab:
		return filter(termRunes, browser.TabMap.GetForSearch())
	case mediatype.Application:
		return filter(termRunes, applications.AppMap.GetForSearch())
	case mediatype.Mimetype:
		return filter(termRunes, applications.MimeMap.GetForSearch())
	case mediatype.Bookmark:
		return filter(termRunes, browser.BookmarkMap.GetForSearch())
	case mediatype.Device:
		return filter(termRunes, power.DeviceMap.GetForSearch())
	case mediatype.File:
		return filter(termRunes, file.FileMap.GetForSearch())
	case mediatype.IconTheme:
		return filter(termRunes, icons.ThemeMap.GetForSearch())
	case mediatype.Start:
		return filter(termRunes, desktopactions.PowerActions.GetForSearch())
	default:
		return []Ranked{}
	}
}

func filter(termRunes []rune, bases []entity.Base) []Ranked {
	var result = make([]Ranked, 0, len(bases))
	for _, res := range bases {
		var rank = match(res.Title, termRunes, 0)
		for _, keyword := range res.Keywords {
			if tmp := match(keyword, termRunes, 0); tmp < rank {
				rank = tmp
			}
		}
		if res.MediaType == mediatype.Start || res.MediaType == mediatype.Application {
			for _, act := range res.Actions {
				if tmp := match(act.Name, termRunes, 0); tmp < rank {
					rank = tmp
				}
			}
		}
		if rank < maxRank {
			result = append(result, Ranked{Base: res, Rank: rank})
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

func Search(term string) []Ranked {
	var termRunes = []rune(strings.ToLower(term))
	var result = make([]Ranked, 0, 1000)

	result = append(result, SearchType(termRunes, mediatype.Notification)...)
	result = append(result, SearchType(termRunes, mediatype.Window)...)
	result = append(result, SearchType(termRunes, mediatype.Tab)...)

	if len(termRunes) > 0 {
		result = append(result, SearchType(termRunes, mediatype.Application)...)
	}
	if len(termRunes) > 2 {
		result = append(result, SearchType(termRunes, mediatype.Device)...)
		result = append(result, SearchType(termRunes, mediatype.File)...)
		result = append(result, SearchType(termRunes, mediatype.Bookmark)...)
		result = append(result, SearchType(termRunes, mediatype.Start)...)
	}

	sort(result)
	return result
}

func SearchByPath(path string) (entity.Base, bool) {
	var bases []entity.Base
	if strings.HasPrefix(path, "/window/") {
		bases = wayland.WindowMap.GetForSearch()
	} else if strings.HasPrefix(path, "/application/") {
		bases = applications.AppMap.GetForSearch()
	} else if strings.HasPrefix(path, "/mimetype/") {
		bases = applications.MimeMap.GetForSearch()
	} else if strings.HasPrefix(path, "/notification/") {
		bases = notifications.NotificationMap.GetForSearch()
	} else if strings.HasPrefix(path, "/icontheme/") {
		bases = icons.ThemeMap.GetForSearch()
	} else if strings.HasPrefix(path, "/device/") {
		bases = power.DeviceMap.GetForSearch()
	} else if strings.HasPrefix(path, "/tab/") {
		bases = browser.TabMap.GetForSearch()
	} else if strings.HasPrefix(path, "/bookmark/") {
		bases = browser.BookmarkMap.GetForSearch()
	} else if strings.HasPrefix(path, "/file/") {
		bases = file.FileMap.GetForSearch()
	} else if strings.HasPrefix(path, "/start/") {
		bases = desktopactions.PowerActions.GetForSearch()
	}

	for _, b := range bases {
		if b.Path == path {
			return b, true
		}
	}

	return entity.Base{}, false
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
			// Add a cost when match not at start or match has skips (TODO: consider how much)
			if j == 0 {
				rnk += 5 * uint(i-lastPos-1)
			} else {
				rnk += 10 * uint(i-lastPos-1)
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
