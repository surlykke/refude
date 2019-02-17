package statusnotifications

import (
	"github.com/surlykke/RefudeServices/lib/resource"
)

func Run(itemCollection *ItemCollection) {
	getOnTheBus()
	go monitorSignals()

	for event := range events {
		//fmt.Println("Got event: ", event)
		var key= event.sender + string(event.path)
		switch event.eventType {
		case ItemCreated:
			item := MakeItem(event.sender, event.path)
			item.AbstractResource = resource.MakeAbstractResource(resource.StandardizedPath("/item/"+item.key), ItemMediaType)
			updateItem(item)
			if (item.menuPath != "") {
				item.LinkTo(resource.Standardizef("/itemmenu/%s/%s", item.sender, item.menuPath), resource.SNI_MENU)
			}
			itemCollection.mutex.Lock()
			itemCollection.items[key] = item
			updateWatcherProperties(itemCollection)
			itemCollection.CachingJsonGetter.ClearByPrefixes("/item/"+key, "/items")
			itemCollection.mutex.Unlock()
			go monitorItem(event.sender, event.path)
		case ItemRemoved:
			itemCollection.mutex.Lock()
			delete(itemCollection.items, key)
			updateWatcherProperties(itemCollection)
			itemCollection.CachingJsonGetter.ClearByPrefixes("/item/"+key, "/items")
			itemCollection.mutex.Unlock()
		case ItemUpdated:
			itemCollection.mutex.Lock()
			if item, ok := itemCollection.items[key]; ok {

				itemCollection.mutex.Unlock()
				var copy= *item
				updateItem(&copy)
				itemCollection.mutex.Lock()

				itemCollection.items[key] = &copy
				itemCollection.CachingJsonGetter.ClearByPrefixes("/item/"+key, "/items")
			}
			itemCollection.mutex.Unlock()
		}
	}
}
