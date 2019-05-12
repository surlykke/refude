// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"fmt"
	"time"

	"github.com/surlykke/RefudeServices/lib/resource"
)

var resourceMap = &ItemRepo{*resource.MakeResourceMap("/items")}
var Items = resource.MakeJsonResourceServer(resourceMap)

func Run() {
	getOnTheBus()
	go monitorSignals()
	var started = time.Now().UnixNano()

	for event := range events {
		var elapsed = time.Now().UnixNano() - started
		var secs = elapsed / 1000000000
		var msecs = (elapsed - secs*1000000000) / 1000000
		var mysecs = (elapsed - secs*1000000000 - msecs*100000) / 1000
		var self = itemSelf(event.sender, event.path)
		switch event.eventType {
		case ItemUpdated, ItemCreated:
			var item = buildItem(event.sender, event.path)
			fmt.Printf("%d.%d.%d:Mapping %s\n", secs, msecs, mysecs, item.Self)
			resourceMap.Set(item.Self, item)
		case ItemRemoved:
			resourceMap.Remove(string(self))
		}
	}
}
