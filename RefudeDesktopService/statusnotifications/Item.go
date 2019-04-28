// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/serialize"
	"github.com/surlykke/RefudeServices/lib/slice"
)

const ItemMediaType resource.MediaType = "application/vnd.org.refude.statusnotifieritem+json"

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
	iconThemePath           string
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

func (item *Item) WriteBytes(w io.Writer) {
	item.GenericResource.WriteBytes(w)
	serialize.String(w, item.Id)
	serialize.String(w, item.Category)
	serialize.String(w, item.Status)
	serialize.String(w, item.IconName)
	serialize.String(w, item.IconAccessibleDesc)
	serialize.String(w, item.AttentionIconName)
	serialize.String(w, item.AttentionAccessibleDesc)
	serialize.String(w, item.Title)
}

type ItemCollection struct {
	*resource.GenericResourceCollection
}

func MakeItemCollection() *ItemCollection {
	var ic = &ItemCollection{resource.MakeGenericResourceCollection()}
	ic.InitializeGenericResourceCollection()
	ic.AddCollectionResource("/items", "/item/")
	return ic
}

func (ic *ItemCollection) Get(path string) resource.Resource {
	if !strings.HasPrefix(path, "/itemmenu/") {
		return ic.GenericResourceCollection.Get(path)
	} else {
		var tmp = string(path[len("/itemmenu/"):])
		if slashPos := strings.Index(tmp, "/"); slashPos == -1 {
			return nil
		} else {
			var sender = tmp[0:slashPos]
			var path = tmp[slashPos:]
			fmt.Println("sender, path:", sender, path)

			if menuItems, err := fetchMenu(sender, dbus.ObjectPath(path)); err != nil {
				return nil
			} else {
				var menu = Menu{resource.MakeGenericResource(menuSelf(sender, dbus.ObjectPath(path)), ""), menuItems}
				//menu.LinkTo(item.GetSelf(), resource.Related)
				return &menu
			}
		}
	}
}
