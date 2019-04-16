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
			item.AbstractResource = resource.MakeAbstractResource(self, ItemMediaType)
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
