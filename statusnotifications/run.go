// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package statusnotifications

import (
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/resource"
)

var itemRepo = repo.MakeRepo[*Item]()
var menuRepo = repo.MakeRepo[*Menu]()
var items = make(chan *Item)
var menus = make(chan *Menu)
var itemRemovals = make(chan string)
var menuRemovals = make(chan string)

func Run() {
	go dbusLoop()
	go itemLoop()
	go menuLoop()
}

func itemLoop() {
	var itemRequests = repo.MakeAndRegisterRequestChan()
	for {
		select {
		case req := <-itemRequests:
			itemRepo.DoRequest(req)
		case item := <-items:
			itemRepo.Put(item)
		case path := <-itemRemovals:
			itemRepo.Remove(path)
		}
	}
}

func menuLoop() {
	var menuRequests = repo.MakeAndRegisterRequestChan()
	for {
		select {
		case req := <-menuRequests:
			menuRepo.DoRequest(req)
		case menu := <-menus:
			menuRepo.Put(menu)
		case path := <-menuRemovals:
			menuRepo.Remove(path)
		}
	}
}

func dbusLoop() {
	getOnTheBus()
	go monitorSignals()

	for event := range events {
		var id = pathEscape(event.sender, event.path)
		var itemPath = "/item/" + id
		var menuPath = "/menu/" + id
		if  event.eventName == "ItemRemoved" {
			itemRemovals <- itemPath
			menuRemovals <- menuPath
		} else { // Assume it's ItemCreated or property update 
			 // A bit bruteforce - if it's a propertychange we could just 
			// retrieve that propery. This is simpler, and probably not too bad
			var item = buildItem(event.sender, event.path)
			items <- item	
			if item.MenuPath != "" {
				menus <- (&Menu{
					ResourceData: *resource.MakeBase("/menu/"+pathEscape(event.sender, item.MenuPath), "Menu", "", "", "menu"),
					sender:       event.sender,
					path:         item.MenuPath,
				})
			}
		}
	}
}
