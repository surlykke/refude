// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package statusnotifications

import (
	"strings"

	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/path"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/resource"
)

func Run() {
	getOnTheBus()
	go monitorSignals()

	for event := range events {
		var path = path.Of("/item/" + event.dbusSender + string(event.dbusPath))
		switch event.name {
		case "ItemRemoved":
			if i, ok := repo.RemoveTyped[*Item](path); ok {
				repo.Remove(i.MenuPath)
			}
		case "ItemCreated":
			var item = buildItem(path, event.dbusSender, event.dbusPath)
			if item.MenuDbusPath != "" {
				repo.Put(&Menu{
					ResourceData: *resource.MakeBase(item.MenuPath, "Menu", "", "", mediatype.Menu),
					DbusSender:   event.dbusSender,
					DbusPath:     item.MenuDbusPath,
					SenderApp:    item.SenderApp,
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

func GetLinks(searchTerm string) []resource.Link {
	var result = make([]resource.Link, 0, 10)

	var getLinksFromMenu func(*Menu, []MenuEntry)
	getLinksFromMenu = func(menu *Menu, entries []MenuEntry) {
		for _, entry := range entries {
			if entry.Type == "standard" {
				if len(entry.SubEntries) > 0 {
					getLinksFromMenu(menu, entry.SubEntries)
				} else {
					var path = path.Of(menu.Path, "?id=", entry.Id)
					var comment = menu.SenderApp
					if strings.Index(comment, "tray") == -1 {
						comment = comment + " tray"
					}
					result = append(result, resource.Link{Path: path, Title: entry.Label, Comment: comment, Icon: entry.Icon, Relation: relation.Action})
				}
			}
		}
	}

	for _, itemMenu := range repo.GetList[*Menu]("/menu/") {
		if entries, err := itemMenu.Entries(); err == nil {
			getLinksFromMenu(itemMenu, entries)
		}

	}
	return result
}
