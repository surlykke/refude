// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"fmt"
	"strings"

	"github.com/godbus/dbus"
	"github.com/surlykke/RefudeServices/lib/respond"
)

type Item struct {
	sender                  string
	itemPath                dbus.ObjectPath
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
	useIconPixmap           bool
	useAttentionIconPixmap  bool
	useOverlayIconPixmap    bool
}

func (item *Item) ToStandardFormat() *respond.StandardFormat {
	return &respond.StandardFormat{
		Self:     itemSelf(item.sender, item.itemPath),
		Type:     "status_item",
		Title:    item.Title,
		IconName: item.IconName,
		OnPost:   "Activate",
		Data:     item,
	}
}

func itemSelf(sender string, path dbus.ObjectPath) string {
	return fmt.Sprintf("/item/%s", strings.Replace(sender+string(path), "/", "-", -1))
}

type ItemMap map[string]*Item

/*func menuSelf(sender string, path dbus.ObjectPath) string {
	return fmt.Sprintf("/itemmenu/%s", strings.Replace(sender+string(path), "/", "-", -1))
}*/

/**
 * dbusMethodName one of "Activate", "SecondaryActivate", "ContextMenu"
 */

/*func (item *Item) POST(w http.ResponseWriter, r *http.Request) {

}*/

/*type MenuResource struct {
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

func (mr *MenuResource) GET(w http.ResponseWriter, r *http.Request) {
	var menu = &Menu{Links: resource.MakeLinks(mr.self, "itemmenu")}
	var err error
	if menu.Entries, err = fetchMenu(mr.sender, mr.path); err != nil {
		log.Println("Error retrieving menu for", mr.sender, mr.path, ":", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else if bytes, err := json.Marshal(menu); err != nil {
		log.Println("Error marshalling menu", err)
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
*/
