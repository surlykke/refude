// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"github.com/surlykke/RefudeServices/lib/resource"
)

var Items = resource.MakeGenericResourceCollection()

func Run() {
	Items.Set("/items", Items.MakePrefixCollection("/item/"))
	getOnTheBus()
	go monitorSignals()

	for event := range events {
		var self = itemSelf(event.sender, event.path)
		switch event.eventType {
		case ItemUpdated, ItemCreated:
			var item = buildItem(event.sender, event.path)
			Items.Set(item.Self, resource.MakeJsonResource(item))
		case ItemRemoved:
			Items.Remove(string(self))
		}
	}
}
