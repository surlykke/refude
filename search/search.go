// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package search

import (
	"fmt"
	"regexp"
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

func Search(term string) []Ranked {
	var re, length = buildRegexp(term)
	var result = make([]Ranked, 0, 1000)

	result = append(result, filter(re, length, notifications.NotificationMap.GetForSearch())...)
	result = append(result, filter(re, length, wayland.WindowMap.GetForSearch())...)
	result = append(result, filter(re, length, browser.TabMap.GetForSearch())...)

	if length > 0 {
		result = append(result, filter(re, length, applications.AppMap.GetForSearch())...)
	}
	if length > 2 {
		result = append(result, filter(re, length, power.DeviceMap.GetForSearch())...)
		result = append(result, filter(re, length, file.FileMap.GetForSearch())...)
		result = append(result, filter(re, length, browser.BookmarkMap.GetForSearch())...)
		result = append(result, filter(re, length, desktopactions.PowerActions.GetForSearch())...)
	}

	sort(result)
	return result
}

func buildRegexp(term string) (*regexp.Regexp, int) {
	if term == "" {
		return nil, 0
	}

	term = strings.ToLower(term)

	var pattern = ""
	var split []string = strings.Split(term, "")
	var length = len(split)
	if length == 1 {
		pattern = regexp.QuoteMeta(split[0])
	} else {
		for i, s := range split {
			var escaped = regexp.QuoteMeta(s)
			if i < len(split)-1 {
				pattern = pattern + fmt.Sprintf("%s[^%s]*", escaped, escaped)
			} else {
				pattern = pattern + escaped
			}
		}
	}
	fmt.Println("Searching:", pattern)
	return regexp.MustCompile(pattern), length

}

func filter(re *regexp.Regexp, termLen int, bases []entity.Base) []Ranked {
	var result = make([]Ranked, 0, len(bases))
	for _, res := range bases {
		var rank = match(res.Title, re, termLen)
		for _, keyword := range res.Keywords {
			if tmp := match(keyword, re, termLen) + 20; tmp < rank {
				rank = tmp
			}
		}
		if res.MediaType == mediatype.Start || res.MediaType == mediatype.Application {
			for _, act := range res.Actions {
				if tmp := match(act.Name, re, termLen) + 40; tmp < rank {
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
func match(text string, re *regexp.Regexp, termLen int) uint {
	text = strings.ToLower(text)
	if re == nil {
		return 0
	}

	if loc := re.FindStringIndex(text); loc == nil {
		return maxRank
	} else {
		return uint(loc[0] + 5*((loc[1]-loc[0])-termLen))
	}
}
