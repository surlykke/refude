package statusnotifications

import (
	"github.com/surlykke/RefudeServices/lib/resource"
)

func Run(itemCollection *ItemCollection) {
	getOnTheBus()
	go monitorSignals()

	for event := range events {
		//fmt.Println("Got event: ", event)
		var self = itemSelf(event.sender, event.path)
		switch event.eventType {
		case ItemCreated:
			item := MakeItem(event.sender, event.path)
			item.AbstractResource = resource.MakeAbstractResource(self, ItemMediaType)
			updateItem(item)
			if (item.menuPath != "") {
				item.LinkTo(resource.Standardizef("/itemmenu/%s/%s", item.sender, item.menuPath), resource.SNI_MENU)
			}
			itemCollection.mutex.Lock()
			itemCollection.items[self] = item
			updateWatcherProperties(itemCollection)
			itemCollection.CachingJsonGetter.ClearByPrefixes(string(self), "/items")
			itemCollection.mutex.Unlock()
			go monitorItem(event.sender, event.path)
		case ItemRemoved:
			itemCollection.mutex.Lock()
			delete(itemCollection.items, self)
			updateWatcherProperties(itemCollection)
			itemCollection.CachingJsonGetter.ClearByPrefixes(string(self), "/items")
			itemCollection.mutex.Unlock()
		case ItemUpdated:
			itemCollection.mutex.Lock()
			if item, ok := itemCollection.items[self]; ok {

				itemCollection.mutex.Unlock()
				var copy= *item
				updateItem(&copy)
				itemCollection.mutex.Lock()

				itemCollection.items[self] = &copy
				itemCollection.CachingJsonGetter.ClearByPrefixes(string(self), "/items")
			}
			itemCollection.mutex.Unlock()
		}
	}
}
