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
)

const ItemMediaType resource.MediaType = "application/vnd.org.refude.statusnotifieritem+json"

type Item struct {
	resource.AbstractResource
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

type ItemCollection struct {
	mutex sync.Mutex
	items map[string]*Item
	server.JsonResponseCache2
	server.PostNotAllowed
	server.PatchNotAllowed
	server.DeleteNotAllowed
}

func (ic *ItemCollection) HandledPrefixes() []string {
	return []string{"/item/", "/items"}
}

func MakeItemCollection() *ItemCollection {
	var itemCollection = &ItemCollection{}
	itemCollection.JsonResponseCache2 = server.MakeJsonResponseCache2(itemCollection)
	itemCollection.items = make(map[string]*Item)
	return itemCollection
}

func (ic *ItemCollection) POST(w http.ResponseWriter, r *http.Request) {
	if res, err := ic.GetResource(r); err != nil {
		requests.ReportUnprocessableEntity(w, err)
	} else if res == nil {
		w.WriteHeader(http.StatusNotFound)
	} else if notification, ok := res.(*Item); !ok {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		action := requests.GetSingleQueryParameter(r, "action", "left")
		x, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "x", "0"))
		y, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "y", "0"))

		fmt.Printf("action: '%s', x: %d, y: %d", action, x, y)
		if slice.Among(action, "left", "middle", "right") {
			action2method := map[string]string{"left": "Activate", "middle": "SecondaryActivate", "right": "ContextMenu"}
			fmt.Println("Calling: ", "org.kde.StatusNotifierItem."+action2method[action], dbus.Flags(0), x, y)
			dbusObj := conn.Object(notification.sender, notification.itemPath)
			call := dbusObj.Call("org.kde.StatusNotifierItem."+action2method[action], dbus.Flags(0), x, y);
			if call.Err != nil {
				log.Println(call.Err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusAccepted)
			}

		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

	}
}

func (ic *ItemCollection) findByMenupath(menupath string) *Item {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	for _, item := range ic.items {
		for _, link := range item.Links {
			if resource.SNI_MENU == link.Rel && menupath == string(link.Href) {
				return item
			}
		}
	}
	return nil
}

func (ic *ItemCollection) GetResource(r *http.Request) (interface{}, error) {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

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
