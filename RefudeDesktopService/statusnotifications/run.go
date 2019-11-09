// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"fmt"

	"github.com/surlykke/RefudeServices/lib/resource"
)

var items = make(map[string]*Item)

func Run() {
	getOnTheBus()
	go monitorSignals()

	// TODO After a restart, pick up those that where?

	updateCollections()
	for event := range events {
		switch event.eventName {
		case "ItemCreated":
			var item = buildItem(event.sender, event.path)
			items[item.Self] = item
		case "ItemRemoved":
			delete(items, itemSelf(event.sender, event.path))
		default:
			var path = itemSelf(event.sender, event.path)
			if item, ok := items[path]; ok {
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
				items[path] = itemCopy
			} else {
				fmt.Println("Item event on unknown item: ", event.sender, event.path)
				continue
			}
		}
		updateCollections()
	}
}

func updateCollections() {
	var resources = make(map[string]interface{}, 2*len(items)+2)
	for _, item := range items {
		resources[item.Self] = item
	}
	resources["/items"] = resource.ExtractResourceList(resources)

	for _, item := range items {
		if item.menu != nil {
			resources[menuSelf(item.menu.sender, item.menu.path)] = item.menu
		}
	}

	resource.MapCollection(&resources, "items")
}
