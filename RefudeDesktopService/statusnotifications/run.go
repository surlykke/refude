// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/ss_events"
	"github.com/surlykke/RefudeServices/lib/searchutils"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/slice"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/items" {
		respond.AsJson(w, r, Collect(searchutils.Term(r)))
	} else if item := get(r.URL.Path); item != nil {
		if r.Method == "POST" {
			dbusObj := conn.Object(item.sender, item.itemPath)
			action := requests.GetSingleQueryParameter(r, "action", "Activate")
			x, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "x", "0"))
			y, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "y", "0"))

			if !slice.Among(action, "Activate", "SecondaryActivate", "ContextMenu") {
				respond.UnprocessableEntity(w, fmt.Errorf("action must be 'Activate', 'SecondaryActivate' or 'ContextMenu'"))
			} else {
				var call = dbusObj.Call("org.kde.StatusNotifierItem."+action, dbus.Flags(0), x, y)
				if call.Err != nil {
					respond.ServerError(w, call.Err)
				} else {
					respond.Accepted(w)
				}
			}
		} else {
			respond.AsJson(w, r, item.ToStandardFormat())
		}
	} else if item = getItemForMenu(r.URL.Path); item != nil {
		if menuItems, err := item.menu(); err != nil {
			respond.ServerError(w, err)
		} else if menuItems != nil {
			if r.Method == "POST" {
				id := requests.GetSingleQueryParameter(r, "id", "")
				idAsInt, _ := strconv.Atoi(id)
				data := dbus.MakeVariant("")
				time := uint32(time.Now().Unix())
				dbusObj := conn.Object(item.sender, item.menuPath)
				call := dbusObj.Call("com.canonical.dbusmenu.Event", dbus.Flags(0), idAsInt, "clicked", data, time)
				if call.Err != nil {
					respond.ServerError(w, err)
				} else {
					respond.Accepted(w)
				}
			} else {
				respond.AsJson(w, r, menuToStandardFormat(r.URL.Path, menuItems))
			}
		} else {
			respond.NotFound(w)
		}
	} else {
		respond.NotFound(w)
	}
}

func getItemForMenu(path string) *Item {
	var item *Item = nil
	if strings.HasSuffix(path, "/menu") {
		item = get(path[0 : len(path)-5])
	}
	if item != nil && item.Menu != "" {
		return item
	} else {
		return nil
	}
}

func Collect(term string) respond.StandardFormatList {
	lock.Lock()
	defer lock.Unlock()
	var sfl = make(respond.StandardFormatList, 0, len(items))
	for _, item := range items {
		if rank := searchutils.SimpleRank(item.Title, "", term); rank > -1 {
			sfl = append(sfl, item.ToStandardFormat().Ranked(rank))
		}
	}
	return sfl.SortByRank()
}

func AllPaths() []string {
	lock.Lock()
	defer lock.Unlock()
	var paths = make([]string, 0, len(items)+1)
	for path, _ := range items {
		paths = append(paths, path)
	}
	paths = append(paths, "/items")
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
}

func sendEvent(path string) {
	ss_events.Publish <- &ss_events.Event{Type: "status_item", Path: path}
}
