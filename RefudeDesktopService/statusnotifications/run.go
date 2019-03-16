package statusnotifications

import (
	"github.com/surlykke/RefudeServices/lib/resource"
	"net/http"
	"strings"
)

var Items = MakeItemCollection()
var Menus = MakeMenuCollection()

func Serve(w http.ResponseWriter, r *http.Request) bool {
	if strings.HasPrefix(r.URL.Path, "/itemmenu") {
		if r.Method == "GET" {
			Menus.GET(w, r)
		} else if r.Method == "POST" {
			Menus.POST(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

		return true
	} else if strings.HasPrefix(r.URL.Path, "/item") {
		if r.Method == "GET" {
			Items.GET(w, r)
		} else if r.Method == "POST" {
			Items.POST(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

		return true
	} else {
		return false
	}
}

func Run() {
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
			Items.mutex.Lock()
			Items.items[self] = item
			updateWatcherProperties(Items)
			Items.CachingJsonGetter.ClearByPrefixes(string(self), "/items")
			Items.mutex.Unlock()
			go monitorItem(event.sender, event.path)
		case ItemRemoved:
			Items.mutex.Lock()
			delete(Items.items, self)
			updateWatcherProperties(Items)
			Items.CachingJsonGetter.ClearByPrefixes(string(self), "/items")
			Items.mutex.Unlock()
		case ItemUpdated:
			Items.mutex.Lock()
			if item, ok := Items.items[self]; ok {

				Items.mutex.Unlock()
				var copy = *item
				updateItem(&copy)
				Items.mutex.Lock()

				Items.items[self] = &copy
				Items.CachingJsonGetter.ClearByPrefixes(string(self), "/items")
			}
			Items.mutex.Unlock()
		}
	}
}
