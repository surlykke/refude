// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package search

import (
	"slices"
	"strings"

	"github.com/surlykke/refude/internal/applications"
	"github.com/surlykke/refude/internal/browser"
	"github.com/surlykke/refude/internal/desktopactions"
	"github.com/surlykke/refude/internal/file"
	"github.com/surlykke/refude/internal/icons"
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/notifications"
	"github.com/surlykke/refude/internal/power"
	"github.com/surlykke/refude/internal/wayland"
	"github.com/surlykke/refude/pkg/bind"
)

const maxRank uint = 1000000

func GetHandler(term string) bind.Response {
	return bind.Json(Search(term))
}

type Ranked struct {
	entity.Base
	Rank uint `json:"-"`
}

func Search(term string) []Ranked {
	var m = makeMatcher(term)
	var result = make([]Ranked, 0, 1000)

	result = append(result, filter(notifications.NotificationMap.GetForSearch(), m)...)
	result = append(result, filter(wayland.WindowMap.GetForSearch(), m)...)
	result = append(result, filter(browser.TabMap.GetForSearch(), m)...)

	if len(m.term) > 0 {
		result = append(result, filter(applications.AppMap.GetForSearch(), m)...)
	}
	if len(m.term) > 2 {
		result = append(result, filter(power.DeviceMap.GetForSearch(), m)...)
		result = append(result, filter(file.FileMap.GetForSearch(), m)...)
		result = append(result, filter(browser.BookmarkMap.GetForSearch(), m)...)
		result = append(result, filter(desktopactions.PowerActions.GetForSearch(), m)...)
	}

	sort(result)
	return result
}

func filter(bases []entity.Base, m matcher) []Ranked {
	var result = make([]Ranked, 0, len(bases))
	for _, res := range bases {
		var rankCalculated = m.match(res.Title)
		for _, keyword := range res.Meta.Keywords {
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
		if b.Meta.Path == path {
			return b, true
		}
	}

	return entity.Base{}, false
}
