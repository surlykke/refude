// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"net/http"
	"github.com/godbus/dbus"
	"log"
	"strconv"
	"fmt"
	"github.com/surlykke/RefudeServices/lib"
	time2 "time"
)

const ItemMediaType lib.MediaType = "application/vnd.org.refude.statusnotifieritem+json"

type Item struct {
	lib.AbstractResource
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

func (item *Item) POST(w http.ResponseWriter, r *http.Request) {
	fmt.Println("POST: ", r.URL)
	action := lib.GetSingleQueryParameter(r, "action", "left")
	x, _ := strconv.Atoi(lib.GetSingleQueryParameter(r, "x", "0"))
	y, _ := strconv.Atoi(lib.GetSingleQueryParameter(r, "y", "0"))
	id := lib.GetSingleQueryParameter(r, "id", "")

	fmt.Println("action: ", action, ", known ids: ", item.menuIds)
	var call *dbus.Call
	if lib.Among(action, "left", "middle", "right") {
		action2method := map[string]string{"left": "Activate", "middle": "SecondaryActivate", "right": "ContextMenu"}
		fmt.Println("Calling: ", "org.kde.StatusNotifierItem."+action2method[action], dbus.Flags(0), x, y)
		dbusObj := conn.Object(item.sender, item.itemPath)
		call = dbusObj.Call("org.kde.StatusNotifierItem."+action2method[action], dbus.Flags(0), x, y);
	} else if action == "menu" && lib.Among(id, item.menuIds...) {
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


