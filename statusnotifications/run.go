// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/watch"

	"github.com/surlykke/RefudeServices/lib/resource"
)

func Run() {
	getOnTheBus()
	go monitorSignals()

	// TODO After a restart, pick up those that where?

	for event := range events {
		var path = "/item/" + pathEscape(event.sender, event.path)
		var menuPath = "/itemmenu/" + pathEscape(event.sender, event.path)
		switch event.eventName {
		case "ItemCreated":
			var item = buildItem(event.sender, event.path)
			Items.Put(resource.MakeResource(path, item.Title, "", item.IconName, "item", item))
			if item.MenuPath != "" {
				Menus.Put(resource.MakeResource(menuPath, "Menu", "", "", "menu", &Menu{event.sender, item.MenuPath}))
			}
		case "ItemRemoved":
			Items.Delete(path)
			Menus.Delete(menuPath)
		default:
			if res := Items.Get(path); res != nil {
				var itemCopy = *(res.Data.(*Item))
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
				Items.Put(resource.MakeResource(path, itemCopy.Title, "", itemCopy.IconName, "item", &itemCopy))
			} else {
				continue
			}

		}
		watch.SomethingChanged("/item/list")
	}
}

var Items = resource.MakeList("/item/list")
var Menus = resource.MakeList("/menuitem/list")

func sendEvent(path string) {
	watch.SomethingChanged(path)
}
