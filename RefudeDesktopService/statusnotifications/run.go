// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/surlykke/RefudeServices/lib/searchutils"

	"github.com/godbus/dbus"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/slice"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if item := get(r.URL.Path); item != nil {
		if r.Method == "GET" {
			respond.AsJson(w, item.ToStandardFormat())
		} else if r.Method == "POST" {
			dbusObj := conn.Object(item.sender, item.itemPath)
			action := requests.GetSingleQueryParameter(r, "action", "Activate")
			x, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "x", "0"))
			y, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "y", "0"))

			if !slice.Among(action, "Activate", "SecondaryActivate", "ContextMenu") {
				respond.UnprocessableEntity(w, fmt.Errorf("action must be 'Activate', 'SecondaryActivate' or 'ContextMenu'"))
			} else {
				var call = dbusObj.Call("org.kde.StatusNotifierItem."+action, dbus.Flags(0), x, y /*FIXME*/)
				if call.Err != nil {
					log.Println(call.Err)
					respond.ServerError(w)
				} else {
					respond.Accepted(w)
				}
			}
		} else {
			respond.NotAllowed(w)
		}
	} else {
		respond.NotFound(w)
	}
}

func SearchItems(collector *searchutils.Collector) {
	lock.Lock()
	defer lock.Unlock()
	for _, item := range items {
		collector.Collect(item.ToStandardFormat())
	}
}

func AllPaths() []string {
	lock.Lock()
	defer lock.Unlock()
	var paths = make([]string, 0, len(items))
	for path, _ := range items {
		paths = append(paths, path)
	}
	return paths
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
				var itemCopy = &(*item)
				switch event.eventName {
				case "org.kde.StatusNotifierItem.NewTitle":
					updateTitle(itemCopy)
				case "org.kde.StatusNotifierItem.NewStatus":
					updateStatus(itemCopy)
				case "org.kde.StatusNotifierItem.NewToolTip":
					updateToolTip(itemCopy)
				case "org.kde.StatusNotifierItem.NewIcon":
					updateIcon(itemCopy)
				case "org.kde.StatusNotifierItem.NewAttentionIcon":
					updateAttentionIcon(itemCopy)
				case "org.kde.StatusNotifierItem.NewOverlayIcon":
					updateOverlayIcon(itemCopy)
				default:
					continue
				}
				set(itemCopy)
			} else {
				fmt.Println("Item event on unknown item: ", event.sender, event.path)
				continue
			}
		}
	}
}

var items = make(ItemMap)
var lock sync.Mutex

func set(item *Item) {
	lock.Lock()
	defer lock.Unlock()
	items[itemSelf(item.sender, item.itemPath)] = item
}

func get(path string) *Item {
	lock.Lock()
	defer lock.Unlock()
	return items[path]
}

func remove(sender string, itemPath dbus.ObjectPath) {
	lock.Lock()
	defer lock.Unlock()
	delete(items, itemSelf(sender, itemPath))
}
