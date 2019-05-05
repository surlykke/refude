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
	"strconv"
	"time"

	"github.com/godbus/dbus"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/slice"
)

const ItemMediaType resource.MediaType = "application/vnd.org.refude.statusnotifieritem+json"

type Item struct {
	resource.GeneralTraits
	resource.DefaultMethods
	key                     string
	sender                  string
	itemPath                dbus.ObjectPath
	menuPath                dbus.ObjectPath
	Id                      string
	Menu                    string
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

func itemSelf(sender string, path dbus.ObjectPath) string {
	return fmt.Sprintf("/item/%s%s", sender, path)
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
