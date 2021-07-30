// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"net/http"
	"regexp"
	"sync"

	"github.com/surlykke/RefudeServices/watch"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

var itemPathPattern = regexp.MustCompile("^(/item/[^/]+)(/menu)?")

func GetJsonResource(r *http.Request) respond.JsonResource {
	if r.URL.Path == "/items" {
		var res = respond.MakeResource("/items", "items", "", "items")
		lock.Lock()
		for _, item := range items {
			res.Links = append(res.Links, item.GetRelatedLink())
		}
		lock.Unlock()
		return &res
	} else if match := itemPathPattern.FindStringSubmatch(r.URL.Path); match != nil {
		if item := get(match[1]); item != nil {
			if match[2] == "/menu" {
				if item.MenuPath != "" {
					return item.buildMenu()
				}
			} else {
				return item
			}
		}
	}
	return nil

}

func Crawl(term string, forDisplay bool, crawler searchutils.Crawler) {
	lock.Lock()
	defer lock.Unlock()
	for _, item := range items {
		crawler(&item.Resource, nil)
	}
}

func Run() {
	getOnTheBus()
	go monitorSignals()

	// TODO After a restart, pick up those that where?

	for event := range events {
		switch event.eventName {
		case "ItemCreated":
			set(buildItem(event.sender, event.path))
		case "ItemRemoved":
			remove(event.sender, event.path)
		default:
			var path = itemSelf(event.sender, event.path)
			if item := get(path); item != nil {
				var tmp = *item
				var itemCopy = &tmp
				switch event.eventName {
				case "org.kde.StatusNotifierItem.NewTitle":
					updateTitle(itemCopy)
				case "org.kde.StatusNotifierItem.NewStatus":
					updateStatus(itemCopy)
				case "org.kde.StatusNotifierItem.NewToolTip":
					updateToolTip(itemCopy)
				case "org.kde.StatusNotifierItem.NewIcon":
					updateIcon(itemCopy)
				case "org.kde.StatusNotifierItem.NewIconThemePath":
					updateIconThemePath(itemCopy)
				case "org.kde.StatusNotifierItem.NewAttentionIcon":
					updateAttentionIcon(itemCopy)
				case "org.kde.StatusNotifierItem.NewOverlayIcon":
					updateOverlayIcon(itemCopy)
				default:
					continue
				}
				set(itemCopy)
			} else {
				continue
			}
		}
	}
}

var items = make(ItemMap)
var lock sync.Mutex

func set(item *Item) {
	var self = itemSelf(item.sender, item.itemPath)
	lock.Lock()
	items[self] = item
	lock.Unlock()
	sendEvent(self)
	sendEvent("/items")
}

func get(path string) *Item {
	lock.Lock()
	defer lock.Unlock()
	return items[path]
}

func remove(sender string, itemPath dbus.ObjectPath) {
	var self = itemSelf(sender, itemPath)
	lock.Lock()
	delete(items, self)
	lock.Unlock()
	sendEvent(self)
	sendEvent("/items")
}

func sendEvent(path string) {
	watch.SomethingChanged(path)
}
