// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"fmt"
	"sort"

	"github.com/surlykke/RefudeServices/lib/resource"
)

var items = make(map[string]*Item)
var resourceMap = resource.MakeResourceMap()
var Items = resource.MakeServer(resourceMap)

func Run() {
	getOnTheBus()
	go monitorSignals()

	// TODO After a restart, pick up those that where?

	updateCollections()
	for event := range events {
		fmt.Println("Event: ", event)
		switch event.eventName {
		case "ItemCreated":
			var item = buildItem(event.sender, event.path)
			items[item.Self] = item
			resourceMap.Set(item.Self, resource.MakeJsonResouceWithEtag(item))
			if item.Menu != "" {
				resourceMap.Set(item.Menu, item.menu)
			}
		case "ItemRemoved":
			if item, ok := items[itemSelf(event.sender, event.path)]; ok {
				resourceMap.Remove(item.Self)
				delete(items, item.Self)
				if item.Menu != "" {
					resourceMap.Remove(item.Menu)
				}
			}
		default:
			if item, ok := items[itemSelf(event.sender, event.path)]; ok {
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
				resourceMap.Set(itemCopy.Self, resource.MakeJsonResouceWithEtag(itemCopy))
			} else {
				fmt.Println("Item event on unknown item: ", event.sender, event.path)

			}
		}
		updateCollections()
		resourceMap.Broadcast()
	}
}

func updateCollections() {
	var list = make(resource.Selfielist, 0, len(items))
	for _, item := range items {
		list = append(list, item)
	}
	sort.Sort(list)
	resourceMap.Set("/items", resource.MakeJsonResouceWithEtag(list))
	resourceMap.Set("/items/brief", resource.MakeJsonResouceWithEtag(list.GetSelfs()))
}
