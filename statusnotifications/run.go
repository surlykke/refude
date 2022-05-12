// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"fmt"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/watch"

	"github.com/surlykke/RefudeServices/lib/resource"
)

func Run() {
	getOnTheBus()
	go monitorSignals()

	// TODO After a restart, pick up those that where?

	for event := range events {
		var id = pathEscape(event.sender, event.path)
		switch event.eventName {
		case "ItemCreated":
			var item = buildItem(event.sender, event.path)
			Items.Put(item)
			if item.MenuPath != "" {
				Menus.Put(&Menu{event.sender, item.MenuPath})
			}
		case "ItemRemoved":
			Items.Delete(id)
			Menus.Delete(id)
		default:
			if res := Items.Get(id); res != nil {
				var itemCopy = *res
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
							itemCopy.IconName = collectPixMap(v)
						}
					} else {
						if v, ok := getProp(itemCopy.sender, itemCopy.path, "IconName"); ok {
							itemCopy.IconName = getStringOr(v)
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
				Items.Put(&itemCopy)
				watch.SomethingChanged(fmt.Sprint("/item/", itemCopy.Id()))
			} else {
				continue
			}

		}
		watch.SomethingChanged("/item/")
	}
}

var Items = resource.MakeCollection[string, *Item]("/item/")
var Menus = resource.MakeCollection[string, *Menu]("/itemmenu/")
