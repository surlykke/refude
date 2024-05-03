// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package statusnotifications

import (
	"github.com/surlykke/RefudeServices/icons"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
)

func Run() {
	getOnTheBus()
	go monitorSignals()

	// TODO After a restart, pick up those that where?

	for event := range events {
		var id = pathEscape(event.sender, event.path)
		var itemPath = "/item/" + id
		var menuPath = "/menu/" + id
		switch event.eventName {
		case "ItemCreated":
			var item = buildItem(event.sender, event.path)
			resourcerepo.Put(item)
			if item.MenuPath != "" {
				resourcerepo.Put(&Menu{
					ResourceData: *resource.MakeBase("/menu/" + pathEscape(event.sender, item.MenuPath), "Menu", "", "", "menu"),
					sender: event.sender, 
					path: item.MenuPath,
				})
			}
		case "ItemRemoved":
			resourcerepo.Remove(itemPath)
			resourcerepo.Remove(menuPath)
		default:
			if item, ok := resourcerepo.GetTyped[*Item](itemPath); ok {
				var itemCopy = *item
				switch event.eventName {
				case "org.kde.StatusNotifierItem.NewTitle":
					if v, ok := getProp(itemCopy.sender, itemCopy.path, "Title"); ok {
						itemCopy.Title = getStringOr(v)
					}
				case "org.kde.StatusNotifierItem.NewStatus":
					if v, ok := getProp(itemCopy.sender, itemCopy.path, "Status"); ok {
						itemCopy.Status = getStringOr(v)
					}
				case "org.kde.StatusNotifierItem.NewToolTip":
					if v, ok := getProp(itemCopy.sender, itemCopy.path, "ToolTip"); ok {
						itemCopy.ToolTip = getStringOr(v)
					}
				case "org.kde.StatusNotifierItem.NewIcon":
					if itemCopy.UseIconPixmap {
						if v, ok := getProp(itemCopy.sender, itemCopy.path, "IconPixmap"); ok {
							itemCopy.IconUrl = link.IconUrlFromName(collectPixMap(v))
						}
					} else {
						if v, ok := getProp(itemCopy.sender, itemCopy.path, "IconName"); ok {
							itemCopy.IconUrl = link.IconUrlFromName(getStringOr(v))
						}
					}
				case "org.kde.StatusNotifierItem.NewIconThemePath":
					if v, ok := getProp(itemCopy.sender, itemCopy.path, "IconThemePath"); ok {
						itemCopy.IconThemePath = getStringOr(v)
						icons.AddBasedir(itemCopy.IconThemePath)
					}
				case "org.kde.StatusNotifierItem.NewAttentionIcon":
					if itemCopy.UseAttentionIconPixmap {
						if v, ok := getProp(itemCopy.sender, itemCopy.path, "AttentionIconPixmap"); ok {
							itemCopy.AttentionIconName = collectPixMap(v)
						}
					} else {
						if v, ok := getProp(itemCopy.sender, itemCopy.path, "AttentionIconName"); ok {
							itemCopy.AttentionIconName = getStringOr(v)
						}
					}
				case "org.kde.StatusNotifierItem.NewOverlayIcon":
					// TODO
				default:
					continue
				}
				resourcerepo.Put(&itemCopy)
			} else {
				continue
			}

		}
	}
}
