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

func Run() {
	getOnTheBus()
	go monitorSignals()

	for event := range events {
		var self = itemSelf(event.sender, event.path)
		switch event.eventType {
		case ItemCreated:
			item := MakeItem(event.sender, event.path)
			item.GenericResource = resource.MakeGenericResource(self, ItemMediaType)
			updateItem(item)
			setItem(item)
		case ItemRemoved:
			removeItem(self)
		case ItemUpdated:
			if item := GetItem(self); item != nil {
				var copy = *item
				updateItem(&copy)
				setItem(&copy)
			}
		}
	}
}
