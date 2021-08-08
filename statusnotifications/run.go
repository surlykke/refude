// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"sync"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/watch"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/resource"
)

func GetResource(pathElements []string) resource.Resource {
	if len(pathElements) == 1 {
		if pathElements[0] == "list" {
			var collection = resource.Collection{resource.MakeLink("/item/list", "Items", "", relation.Self)}

			for _, item := range items {
				collection = append(collection, resource.MakeLink(item.self, item.Title, item.IconName, relation.Related))
			}

			return collection
		} else if item := get("/item/" + pathElements[0]); item != nil {
			return item
		}
	} else if len(pathElements) == 2 && pathElements[1] == "menu" {
		if item := get("/item/" + pathElements[0]); item != nil {
			if menu := item.buildMenu(); menu != nil {
				return menu
			}
		}
	}
	return nil
}

func CollectPaths(method string, sink chan string) {
	lock.Lock()
	defer lock.Unlock()
	sink <- "/item/list"
	for _, item := range items {
		sink <- item.self
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
				var itemCopy = *item
				switch event.eventName {
				case "org.kde.StatusNotifierItem.NewTitle":
					if v, ok := getProp(itemCopy.sender, itemCopy.itemPath, "Title"); ok {
						itemCopy.Title = getStringOr(v)
					}
				case "org.kde.StatusNotifierItem.NewStatus":
					if v, ok := getProp(itemCopy.sender, itemCopy.itemPath, "Status"); ok {
						itemCopy.Status = getStringOr(v)
					}
				case "org.kde.StatusNotifierItem.NewToolTip":
					if v, ok := getProp(itemCopy.sender, itemCopy.itemPath, "ToolTip"); ok {
						itemCopy.ToolTip = getStringOr(v)
					}
				case "org.kde.StatusNotifierItem.NewIcon":
					if itemCopy.UseIconPixmap {
						if v, ok := getProp(itemCopy.sender, itemCopy.itemPath, "IconPixmap"); ok {
							itemCopy.IconName = collectPixMap(v)
						}
					} else {
						if v, ok := getProp(itemCopy.sender, itemCopy.itemPath, "IconName"); ok {
							itemCopy.IconName = getStringOr(v)
						}
					}
				case "org.kde.StatusNotifierItem.NewIconThemePath":
					if v, ok := getProp(itemCopy.sender, itemCopy.itemPath, "IconThemePath"); ok {
						itemCopy.IconThemePath = getStringOr(v)
						icons.AddBasedir(itemCopy.IconThemePath)
					}
				case "org.kde.StatusNotifierItem.NewAttentionIcon":
					if itemCopy.UseAttentionIconPixmap {
						if v, ok := getProp(itemCopy.sender, itemCopy.itemPath, "AttentionIconPixmap"); ok {
							itemCopy.AttentionIconName = collectPixMap(v)
						}
					} else {
						if v, ok := getProp(itemCopy.sender, itemCopy.itemPath, "AttentionIconName"); ok {
							itemCopy.AttentionIconName = getStringOr(v)
						}
					}
				case "org.kde.StatusNotifierItem.NewOverlayIcon":
					// TODO
				default:
					continue
				}
				set(&itemCopy)
			} else {
				continue
			}
		}
	}
}

var items = make(ItemMap)
var lock sync.Mutex

func set(item *Item) {
	lock.Lock()
	items[item.self] = item
	lock.Unlock()
	sendEvent(item.self)
	sendEvent("/item/list")
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
	sendEvent("/item/list")
}

func sendEvent(path string) {
	watch.SomethingChanged(path)
}
