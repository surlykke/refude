// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package statusnotifications

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/slice"
)

type MenuEntry struct {
	Id          string
	Type        string
	Label       string
	Enabled     bool
	Visible     bool
	IconUrl     string
	Shortcuts   [][]string `json:",omitempty"`
	ToggleType  string     `json:",omitempty"`
	ToggleState int32
	SubEntries  []MenuEntry `jsoControllern:",omitempty"`
}

type Menu struct {
	resource.ResourceData
	DbusSender string
	DbusPath   dbus.ObjectPath
}

var emptyList = []byte("[]")

func (m *Menu) MarshalJSON() ([]byte, error) {
	if entries, err := m.Entries(); err == nil {
		return respond.ToJson(entries), nil
	} else {
		log.Warn(err)
		return emptyList, nil
	}
}

func (m *Menu) Entries() ([]MenuEntry, error) {
	obj := conn.Object(m.DbusSender, m.DbusPath)
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
	} else {
		if len(menu.SubEntries) > 0 {
			return clean(menu.SubEntries), nil
		} else {
			return []MenuEntry{}, nil
		}
	}
}

func parseMenu(value []interface{}) (MenuEntry, error) {
	var menuItem = MenuEntry{}
	var id int32
	var ok bool
	var m map[string]dbus.Variant
	var s []dbus.Variant

	if len(value) < 3 {
		return MenuEntry{}, errors.New("Wrong length")
	} else if id, ok = value[0].(int32); !ok {
		return MenuEntry{}, errors.New("Expected int32, got: " + reflect.TypeOf(value[0]).String())
	} else if m, ok = value[1].(map[string]dbus.Variant); !ok {
		return MenuEntry{}, errors.New("Excpected dbus.Variant, got: " + reflect.TypeOf(value[1]).String())
	} else if s, ok = value[2].([]dbus.Variant); !ok {
		return MenuEntry{}, errors.New("expected []dbus.Variant, got: " + reflect.TypeOf(value[2]).String())
	}

	menuItem.Id = fmt.Sprintf("%d", id)

	menuItem.Type = getString(m["type"])
	if menuItem.Type == "" {
		menuItem.Type = "standard"
	}
	if !slice.Among(menuItem.Type, "standard", "separator") {
		return MenuEntry{}, errors.New("Illegal menuitem type: " + menuItem.Type)
	}
	menuItem.Label = strings.ReplaceAll(getString(m["label"]), "_", "")
	menuItem.Enabled = getBool(m["enabled"], true)
	menuItem.Visible = getBool(m["visible"], true)
	var iconName = getString(m["icon-name"])
	// TODO: Look for pixmap
	if iconName != "" {
		menuItem.IconUrl = icons.UrlFromName(iconName)
	}

	if menuItem.ToggleType = getString(m["toggle-type"]); !slice.Among(menuItem.ToggleType, "checkmark", "radio", "") {
		return MenuEntry{}, errors.New("Illegal toggle-type: " + menuItem.ToggleType)
	}

	menuItem.ToggleState = getInt32(m["toggle-state"], -1)
	if childrenDisplay := getString(m["children-display"]); childrenDisplay == "submenu" || len(m) == 0 {
		for _, variant := range s {
			if interfaces, ok := variant.Value().([]interface{}); !ok {
				return MenuEntry{}, errors.New("Submenu item not of type []interface")
			} else {
				if submenuItem, err := parseMenu(interfaces); err != nil {
					return MenuEntry{}, err
				} else {
					menuItem.SubEntries = append(menuItem.SubEntries, submenuItem)
				}
			}
		}
	} else if childrenDisplay != "" {
		log.Warn("Ignoring unknown children-display type:", childrenDisplay)
	}

	return menuItem, nil
}

/*
 * Some trays have menus with consecutive separators. We replace any such sequence with a single one.
 */
func clean(menuEntries []MenuEntry) []MenuEntry {
	var cleaned = make([]MenuEntry, 0, len(menuEntries))
	var justSawSeparator = false
	for _, entry := range menuEntries {
		if entry.Type != "separator" || !justSawSeparator {
			entry.SubEntries = clean(entry.SubEntries)
			cleaned = append(cleaned, entry)
		}
		justSawSeparator = entry.Type == "separator"
	}
	return cleaned
}

func (m *Menu) DoPost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("menuPost:", r.URL.Path, r.URL.Query())
	var menuId = requests.GetSingleQueryParameter(r, "id", "")
	if menuId == "" {
		respond.NotFound(w)
	} else {
		idAsInt, _ := strconv.Atoi(menuId)
		data := dbus.MakeVariant("")
		time := uint32(time.Now().Unix())
		dbusObj := conn.Object(m.DbusSender, m.DbusPath)
		var call = dbusObj.Call("com.canonical.dbusmenu.Event", dbus.Flags(0), idAsInt, "clicked", data, time)
		respond.Accepted(w)
		if call.Err != nil {
			log.Warn("Error in call", call.Err)
		} else {
			log.Info("Call succeeded")
		}
	}
}
