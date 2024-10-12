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
		var path = "/item/" + event.dbusSender + string(event.dbusPath)
		switch event.name {
		case "ItemRemoved":
			if i, ok := repo.RemoveTyped[*Item](path); ok {
				repo.Remove("/menu" + i.DbusSender + string(i.MenuDbusPath))
			}
		case "ItemCreated":
			var item = buildItem(path, event.dbusSender, event.dbusPath)
			if item.MenuDbusPath != "" {
				var menuPath = "/menu/" + event.dbusSender + string(item.MenuDbusPath)
				item.AddLink(resource.NormalizeHref(menuPath), "", "", relation.Menu)
				repo.Put(&Menu{
					ResourceData: *resource.MakeBase(menuPath, "Menu", "", "", mediatype.Menu),
					DbusSender:   event.dbusSender,
					DbusPath:     item.MenuDbusPath,
				})
			}
			repo.Put(item)
		case "NewTitle", "NewIcon", "NewAttentionIcon", "NewOverlayIcon", "NewToolTip", "NewStatus":
			if item, ok := repo.Get[*Item](path); ok {
				var copy = *item
				switch event.name {
				case "NewTitle":
				case "NewIcon":
					RetrieveIcon(&copy)
				case "NewAttentionIcon":
					RetrieveAttentionIcon(&copy)
				case "NewOverlayIcon":
					RetrieveOverlayIcon(&copy)
				case "NewToolTip":
					RetrieveToolTip(&copy)
				case "NewStatus":
					RetrieveStatus(&copy)
				}

			}
		}
	}
}
