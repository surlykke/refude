// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/godbus/dbus"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/slice"
)

const ItemMediaType resource.MediaType = "application/vnd.org.refude.statusnotifieritem+json"

var items = make(map[resource.StandardizedPath]*Item)
var lock sync.Mutex

func GetItem(path resource.StandardizedPath) *Item {
	lock.Lock()
	defer lock.Unlock()

	return items[path]
}

func setItem(item *Item) {
	lock.Lock()
	defer lock.Unlock()

	items[item.GetSelf()] = item
	updateWatcherProperties()
}

func removeItem(path resource.StandardizedPath) {
	lock.Lock()
	defer lock.Unlock()

	delete(items, path)
	updateWatcherProperties()
}

func GetItems() []resource.Resource {
	lock.Lock()
	defer lock.Unlock()

	var res = make([]resource.Resource, 0, len(items))
	for _, item := range items {
		res = append(res, item)
	}
	sort.Sort(resource.ResourceCollection(res))
	return res
}

func findByMenupath(menupath string) *Item {
	lock.Lock()
	defer lock.Unlock()

	for _, item := range items {
		for _, link := range item.Links {
			if resource.SNI_MENU == link.Rel && menupath == string(link.Href) {
				return item
			}
		}
	}
	return nil
}

type Item struct {
	resource.GenericResource
	key                     string
	sender                  string
	itemPath                dbus.ObjectPath
	menuPath                dbus.ObjectPath
	Id                      string
	Category                string
	Status                  string
	IconName                string
	IconAccessibleDesc      string
	AttentionIconName       string
	AttentionAccessibleDesc string
	Title                   string
	Menu                    []MenuItem `json:",omitempty"`

	iconThemePath string
}

func MakeItem(sender string, path dbus.ObjectPath) *Item {
	return &Item{key: sender + string(path), sender: sender, itemPath: path}
}

func itemSelf(sender string, path dbus.ObjectPath) resource.StandardizedPath {
	return resource.Standardizef("/item/%s%s", sender, path)
}

func (item *Item) POST(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POST: ", r.URL)
	if item.menuPath == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	action := requests.GetSingleQueryParameter(r, "action", "left")
	x, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "x", "0"))
	y, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "y", "0"))
	id := requests.GetSingleQueryParameter(r, "id", "")

	fmt.Println("action: ", action)
	var call *dbus.Call
	if slice.Among(action, "left", "middle", "right") {
		action2method := map[string]string{"left": "Activate", "middle": "SecondaryActivate", "right": "ContextMenu"}
		fmt.Println("Calling: ", "org.kde.StatusNotifierItem."+action2method[action], dbus.Flags(0), x, y)
		dbusObj := conn.Object(item.sender, item.itemPath)
		call = dbusObj.Call("org.kde.StatusNotifierItem."+action2method[action], dbus.Flags(0), x, y)
	} else if action == "menu" {
		idAsInt, _ := strconv.Atoi(id)
		data := dbus.MakeVariant("")
		time := uint32(time.Now().Unix())
		fmt.Println("Calling: ", "com.canonical.dbusmenu.Event", dbus.Flags(0), idAsInt, "clicked", data, time)
		dbusObj := conn.Object(item.sender, item.menuPath)
		call = dbusObj.Call("com.canonical.dbusmenu.Event", dbus.Flags(0), idAsInt, "clicked", data, time)
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if call.Err != nil {
		log.Println(call.Err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}
