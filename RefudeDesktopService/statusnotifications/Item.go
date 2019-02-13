// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"fmt"
	"github.com/godbus/dbus"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/server"
	"github.com/surlykke/RefudeServices/lib/slice"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	time2 "time"
)

const ItemMediaType resource.MediaType = "application/vnd.org.refude.statusnotifieritem+json"

type Item struct {
	resource.AbstractResource
	key                     string
	Id                      string
	Category                string
	Status                  string
	IconName                string
	IconAccessibleDesc      string
	AttentionIconName       string
	AttentionAccessibleDesc string
	Title                   string
	Menu                    []MenuItem `json:",omitempty"`

	sender   string
	itemPath dbus.ObjectPath
	menuPath dbus.ObjectPath

	iconThemePath string
	path          string
	menuIds       []string
}

type MenuItem struct {
	Id          string
	Type        string
	Label       string
	Enabled     bool
	Visible     bool
	IconName    string
	Shortcuts   [][]string `json:",omitempty"`
	ToggleType  string     `json:",omitempty"`
	ToggleState int32
	SubMenus    []MenuItem `json:",omitempty"`
}

func MakeItem(sender string, path dbus.ObjectPath) *Item {
	return &Item{key: sender + string(path), sender: sender, itemPath: path}
}

func (item *Item) POST(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POST: ", r.URL)
	action := requests.GetSingleQueryParameter(r, "action", "left")
	x, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "x", "0"))
	y, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "y", "0"))
	id := requests.GetSingleQueryParameter(r, "id", "")

	fmt.Println("action: ", action, ", known ids: ", item.menuIds)
	var call *dbus.Call
	if slice.Among(action, "left", "middle", "right") {
		action2method := map[string]string{"left": "Activate", "middle": "SecondaryActivate", "right": "ContextMenu"}
		fmt.Println("Calling: ", "org.kde.StatusNotifierItem."+action2method[action], dbus.Flags(0), x, y)
		dbusObj := conn.Object(item.sender, item.itemPath)
		call = dbusObj.Call("org.kde.StatusNotifierItem."+action2method[action], dbus.Flags(0), x, y);
	} else if action == "menu" && slice.Among(id, item.menuIds...) {
		idAsInt, _ := strconv.Atoi(id)
		data := dbus.MakeVariant("")
		time := uint32(time2.Now().Unix())
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


type ItemCollection struct {
	sync.Mutex
	server.JsonResponseCache
	items map[string]*Item
}

func MakeItemCollection() *ItemCollection {
	var itemCollection = &ItemCollection{}
	itemCollection.JsonResponseCache = server.MakeJsonResponseCache(itemCollection)
	itemCollection.items = make(map[string]*Item)
	return itemCollection
}

func (ic *ItemCollection) findByMenuPath(sender string, menuPath dbus.ObjectPath) *Item {
	for _, item := range ic.items {
		if sender == item.sender && menuPath == item.menuPath {
			return item
		}
	}
	return nil
}


func (ic *ItemCollection) GetResource(r *http.Request) (interface{}, error) {
	var path = r.URL.Path
	if path == "/items" {
		var items = make([]*Item, 0, len(ic.items))

		var matcher, err = requests.GetMatcher(r);
		if err != nil {
			return nil, err
		}

		for _, item := range ic.items {
			if matcher(item) {
				items = append(items, item)
			}
		}

		return items, nil
	} else if strings.HasPrefix(path, "/item/") {
		if item, ok := ic.items[path[len("/item/"):]]; ok {
			return item, nil
		} else {
			return nil, nil
		}
	} else {
		return nil, nil
	}
}
