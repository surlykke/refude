// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

var Items = MakeItemCollection()

func Run() {
	getOnTheBus()
	go monitorSignals()

	for event := range events {
		var self = itemSelf(event.sender, event.path)
		switch event.eventType {
		case ItemUpdated, ItemCreated:
			Items.Set(buildItem(event.sender, event.path))
		case ItemRemoved:
			Items.Remove(string(self))
		}
	}
}
