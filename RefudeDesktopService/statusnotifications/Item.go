// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/slice"
)

type Item struct {
	resource.Links
	key                     string
	sender                  string
	itemPath                dbus.ObjectPath
	menu                    *MenuResource
	Id                      string
	Menu                    string `json:",omitempty"`
	Category                string
	Status                  string
	IconName                string
	IconAccessibleDesc      string
	AttentionIconName       string
	AttentionAccessibleDesc string
	Title                   string
	ToolTip                 string
	iconThemePath           string
}

func MakeItem(sender string, path dbus.ObjectPath) *Item {
	return &Item{
		Links:    resource.Links{Self: itemSelf(sender, path), RefudeType: "statusnotifieritem"},
		key:      sender + string(path),
		sender:   sender,
		itemPath: path,
	}
}

func itemSelf(sender string, path dbus.ObjectPath) string {
	return fmt.Sprintf("/item/%s", strings.Replace(sender+string(path), "/", "-", -1))
}

func menuSelf(sender string, path dbus.ObjectPath) string {
	return fmt.Sprintf("/itemmenu/%s", strings.Replace(sender+string(path), "/", "-", -1))
}

func (item *Item) POST(w http.ResponseWriter, r *http.Request) {
	action := requests.GetSingleQueryParameter(r, "action", "left")
	x, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "x", "0"))
	y, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "y", "0"))

	var call *dbus.Call
	if slice.Among(action, "left", "middle", "right") {
		action2method := map[string]string{"left": "Activate", "middle": "SecondaryActivate", "right": "ContextMenu"}
		dbusObj := conn.Object(item.sender, item.itemPath)
		call = dbusObj.Call("org.kde.StatusNotifierItem."+action2method[action], dbus.Flags(0), x, y)
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

type MenuResource struct {
	self   string
	sender string
	path   dbus.ObjectPath
}

func MakeMenuResource(sender string, path dbus.ObjectPath) *MenuResource {
	return &MenuResource{
		self:   menuSelf(sender, path),
		sender: sender,
		path:   path,
	}
}

func (m *MenuResource) GetSelf() string {
	return m.self
}

func (mr *MenuResource) GET(w http.ResponseWriter, r *http.Request) {
	var menu = &Menu{Links: resource.Links{Self: mr.self, RefudeType: "itemmenu"}}
	var err error
	if menu.Entries, err = fetchMenu(mr.sender, mr.path); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else if bytes, err := json.Marshal(menu); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

func (mr *MenuResource) POST(w http.ResponseWriter, r *http.Request) {
	id := requests.GetSingleQueryParameter(r, "id", "")
	idAsInt, _ := strconv.Atoi(id)
	data := dbus.MakeVariant("")
	time := uint32(time.Now().Unix())
	dbusObj := conn.Object(mr.sender, mr.path)
	call := dbusObj.Call("com.canonical.dbusmenu.Event", dbus.Flags(0), idAsInt, "clicked", data, time)
	if call.Err != nil {
		log.Println(call.Err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

type Menu struct {
	resource.Links
	Entries []MenuItem
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
	SubEntries  []MenuItem `jsoControllern:",omitempty"`
}

func fetchMenu(sender string, path dbus.ObjectPath) ([]MenuItem, error) {
	obj := conn.Object(sender, path)
	if call := obj.Call(MENU_INTERFACE+".GetLayout", dbus.Flags(0), 0, -1, []string{}); call.Err != nil {
		return nil, call.Err
	} else if len(call.Body) < 2 {
		return nil, errors.New(fmt.Sprint("Retrieved", len(call.Body), "arguments, expected 2"))
	} else if _, ok := call.Body[0].(uint32); !ok {
		return nil, errors.New(fmt.Sprint("Expected uint32 as first return argument, got:", reflect.TypeOf(call.Body[0])))
	} else if interfaces, ok := call.Body[1].([]interface{}); !ok {
		return nil, errors.New(fmt.Sprint("Expected []interface{} as second return argument, got:", reflect.TypeOf(call.Body[1])))
	} else if menu, err := parseMenu(interfaces); err != nil {
		return nil, err
	} else if len(menu.SubEntries) > 0 {
		return menu.SubEntries, nil
	} else {
		return []MenuItem{menu}, nil
	}
}
