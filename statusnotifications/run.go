// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package statusnotifications

import (
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/resource"
)

func Run() {
	getOnTheBus()
	go monitorSignals()

	for event := range events {
		var id = pathEscape(event.sender, event.path)
		var itemPath = "/item/" + id
		var menuPath = "/menu/" + id
		if event.eventName == "ItemRemoved" {
			repo.Remove(itemPath)
			repo.Remove(menuPath)
		} else {
			// Assume it's ItemCreated or property update
			// A bit bruteforce - if it's a propertychange we could just
			// retrieve that propery. This is simpler, and probably not too bad
			var item = buildItem(itemPath, event.sender, event.path)
			if item.MenuPath != "" {
				var menuPath = "/menu/" + pathEscape(event.sender, item.MenuPath)
				item.AddLink(menuPath, "", "", relation.Menu)
				repo.Put(&Menu{
					ResourceData: *resource.MakeBase(menuPath, "Menu", "", "", mediatype.Menu),
					sender:       event.sender,
					path:         item.MenuPath,
				})
			}
			repo.Put(item)
		}
	}
}
