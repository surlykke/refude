// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"sort"

	"github.com/godbus/dbus"

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
		switch event.eventType {
		case ItemUpdated:
			var item = buildItem(event.sender, event.path)
			items[item.Self] = item
			resourceMap.Set(item.Self, resource.MakeJsonResouceWithEtag(item))
		case MenuUpdated:
			if itemPath := menuPath2ItemPath(event.sender, event.path); itemPath != "" {
				var item = buildItem(event.sender, itemPath)
				items[item.Self] = item
				resourceMap.Set(item.Self, resource.MakeJsonResouceWithEtag(item))

			}
		case ItemRemoved:
			var self = itemSelf(event.sender, event.path)
			resourceMap.Remove(self)
			delete(items, self)
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

func menuPath2ItemPath(sender string, menuPath dbus.ObjectPath) dbus.ObjectPath {
	for _, item := range items {
		if item.sender == sender && item.menuPath == menuPath {
			return item.itemPath
		}
	}

	return ""
}
