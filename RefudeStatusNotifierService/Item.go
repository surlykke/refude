// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"net/http"
	"github.com/surlykke/RefudeServices/lib/icons"
	"github.com/godbus/dbus"
	"log"
	"strconv"
	"github.com/surlykke/RefudeServices/lib/resource"
	"fmt"
	"reflect"
	"errors"
	"github.com/surlykke/RefudeServices/lib/utils"
	time2 "time"
	"strings"
)

const ItemMediaType resource.MediaType = "application/vnd.org.refude.statusnotifieritem+json"

type Item struct {
	resource.ByteResource
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
	Self          string
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
	action := resource.GetSingleQueryParameter(r, "action", "left")
	x, _ := strconv.Atoi(resource.GetSingleQueryParameter(r, "x", "0"))
	y, _ := strconv.Atoi(resource.GetSingleQueryParameter(r, "y", "0"))
	id := resource.GetSingleQueryParameter(r, "id", "")

	fmt.Println("action: ", action, ", known ids: ", item.menuIds)
	var call *dbus.Call
	if utils.Among(action, "left", "middle", "right") {
		action2method := map[string]string{"left": "Activate", "middle": "SecondaryActivate", "right": "ContextMenu"}
		fmt.Println("Calling: ", "org.kde.StatusNotifierItem."+action2method[action], dbus.Flags(0), x, y)
		dbusObj := conn.Object(item.sender, item.itemPath)
		call = dbusObj.Call("org.kde.StatusNotifierItem."+action2method[action], dbus.Flags(0), x, y);
	} else if action == "menu" && utils.Among(id, item.menuIds...) {
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

func (item *Item) mappableCopy() *Item {
	var tmp = *item
	tmp.SetBytes(resource.ToJSon(tmp))
	return &tmp
}

func getSingleProp(service string, path dbus.ObjectPath, propName string) dbus.Variant {
	if call := conn.Object(service, path).Call(PROPERTIES_INTERFACE+".Get", dbus.Flags(0), ITEM_INTERFACE, propName); call.Err != nil {
		return dbus.Variant{}
	} else {
		return call.Body[0].(dbus.Variant)
	}
}

func getAllProps(sender string, dbusPath dbus.ObjectPath) map[string]dbus.Variant {
	if call := conn.Object(sender, dbusPath).Call(PROPERTIES_INTERFACE+".GetAll", dbus.Flags(0), ITEM_INTERFACE); call.Err != nil {
		return map[string]dbus.Variant{}
	} else {
		return call.Body[0].(map[string]dbus.Variant)
	}
}

func (item *Item) fetchMenu() {
	item.Menu, item.menuIds = []MenuItem{}, []string{}
	obj := conn.Object(item.sender, item.menuPath)
	if call := obj.Call(MENU_INTERFACE+".GetLayout", dbus.Flags(0), 0, -1, []string{}); call.Err != nil {
		log.Println(call.Err)
	} else if len(call.Body) < 2 {
		log.Println("Retrieved", len(call.Body), "arguments, expected 2")
	} else if _, ok := call.Body[0].(uint32); !ok {
		log.Println("Expected uint32 as first return argument, got:", reflect.TypeOf(call.Body[0]))
	} else if interfaces, ok := call.Body[1].([]interface{}); !ok {
		log.Println("Expected []interface{} as second return argument, got:", reflect.TypeOf(call.Body[1]))
	} else if menu, menuIds, err := parseMenu(interfaces); err != nil {
		log.Println("Error retrieving menu", err)
	} else if len(menu.SubMenus) > 0 {
		item.Menu, item.menuIds = menu.SubMenus, utils.Remove(menuIds, menu.Id)
	} else {
		item.Menu, item.menuIds = []MenuItem{menu}, menuIds
	}
}

func getStringOr(m map[string]dbus.Variant, key string, fallback string) string {
	if variant, ok := m[key]; ok {
		if res, ok := variant.Value().(string); !ok {
			log.Println("Looking for string at", key, "got:", reflect.TypeOf(variant.Value()))
		} else {
			return res
		}
	}

	return fallback
}

func getBoolOr(m map[string]dbus.Variant, key string, fallback bool) bool {
	if variant, ok := m[key]; ok {
		if res, ok := variant.Value().(bool); !ok {
			log.Println("Looking for boolean at", key, "got:", reflect.TypeOf(variant.Value()))
		} else {
			return res
		}
	}

	return fallback
}

func getInt32Or(m map[string]dbus.Variant, key string, fallback int32) int32 {
	if variant, ok := m[key]; ok {
		if res, ok := variant.Value().(int32); !ok {
			log.Println("Looking for int at", key, "got:", reflect.TypeOf(variant.Value()))
		} else {
			return res
		}
	}
	return fallback
}

func getDbusPath(m map[string]dbus.Variant, key string) dbus.ObjectPath {
	if variant, ok := m[key]; ok {
		if res, ok := variant.Value().(dbus.ObjectPath); !ok {
			log.Println("Looking for ObjectPath at", key, "got:", reflect.TypeOf(variant.Value()))
		} else {
			return res
		}
	}
	return ""
}

func parseMenu(value []interface{}) (MenuItem, []string, error) {
	var menuItem = MenuItem{}
	var id int32
	var ok bool
	var m map[string]dbus.Variant
	var s []dbus.Variant

	if len(value) < 3 {
		return MenuItem{}, []string{}, errors.New("Wrong length")
	} else if id, ok = value[0].(int32); !ok {
		return MenuItem{}, []string{}, errors.New("Expected int32, got: " + reflect.TypeOf(value[0]).String())
	} else if m, ok = value[1].(map[string]dbus.Variant); !ok {
		return MenuItem{}, []string{}, errors.New("Excpected dbus.Variant, got: " + reflect.TypeOf(value[1]).String())
	} else if s, ok = value[2].([]dbus.Variant); !ok {
		return MenuItem{}, []string{}, errors.New("expected []dbus.Variant, got: " + reflect.TypeOf(value[2]).String())
	}

	menuItem.Id = fmt.Sprintf("%d", id)

	if menuItem.Type = getStringOr(m, "type", "standard");
		!utils.Among(menuItem.Type, "standard", "separator") {
		return MenuItem{}, []string{}, errors.New("Illegal menuitem type: " + menuItem.Type)
	}
	menuItem.Label = getStringOr(m, "label", "")
	menuItem.Enabled = getBoolOr(m, "enabled", true)
	menuItem.Visible = getBoolOr(m, "visible", true)
	if menuItem.IconName = getStringOr(m, "icon-name", ""); menuItem.IconName == "" {
		// FIXME: Look for pixmap
	}
	if menuItem.ToggleType = getStringOr(m, "toggle-type", ""); !utils.Among(menuItem.ToggleType, "checkmark", "radio", "") {
		return MenuItem{}, []string{}, errors.New("Illegal toggle-type: " + menuItem.ToggleType)
	}

	menuItem.ToggleState = getInt32Or(m, "toggle-state", -1)
	var menuIds = []string{menuItem.Id}
	if childrenDisplay := getStringOr(m, "children-display", ""); childrenDisplay == "submenu" {
		for _, variant := range s {
			if interfaces, ok := variant.Value().([]interface{}); !ok {
				return MenuItem{}, []string{}, errors.New("Submenu item not of type []interface")
			} else {
				if submenuItem, subIds, err := parseMenu(interfaces); err != nil {
					return MenuItem{}, []string{}, err
				} else {
					menuIds = append(menuIds, subIds...)
					menuItem.SubMenus = append(menuItem.SubMenus, submenuItem)
				}
			}
		}
	} else if childrenDisplay != "" {
		log.Println("warning: ignoring unknown children-display type:", childrenDisplay)
	}

	return menuItem, menuIds, nil
}

func (item *Item) fetchProps() {
	props := getAllProps(item.sender, item.itemPath)

	item.Id = getStringOr(props, "ID", "")
	item.Category = getStringOr(props, "Category", "")
	item.Status = getStringOr(props, "Status", "")
	item.IconName = getStringOr(props, "IconName", "")
	item.IconAccessibleDesc = getStringOr(props, "IconAccessibleDesc", "")
	item.AttentionIconName = getStringOr(props, "AttentionIconName", "")
	item.AttentionAccessibleDesc = getStringOr(props, "AttentionAccessibleDesc", "")
	item.Title = getStringOr(props, "Title", "")
	item.menuPath = getDbusPath(props, "Menu")
	item.iconThemePath = getStringOr(props, "IconThemePath", "")

	if item.IconName == "" {
		item.IconName = collectPixMap(props, "IconPixmap")
	} else if item.iconThemePath != "" {
		icons.CopyIcons(item.IconName, item.iconThemePath)
	}
	if item.AttentionIconName == "" {
		item.AttentionIconName = collectPixMap(props, "AttentionIconPixmap")
	} else if item.iconThemePath != "" {
		icons.CopyIcons(item.AttentionIconName, item.iconThemePath)
	}
}

func (item *Item) restPath() string {
	return "/items/" + strings.Replace(item.sender[1:]+string(item.itemPath), "/", "-", -1)
}

func collectPixMap(m map[string]dbus.Variant, key string) string {
	if variant, ok := m[key]; ok {
		if arrs, ok := variant.Value().([][]interface{}); !ok {
			log.Println("Looking for [][]interface{} at"+key+", got:", reflect.TypeOf(variant.Value()))
		} else {
			res := make(icons.Icon, 0)
			for _, arr := range (arrs) {
				for len(arr) > 2 {
					width := arr[0].(int32)
					height := arr[1].(int32)
					pixels := arr[2].([]byte)
					res = append(res, icons.Img{Width: width, Height: height, Pixels: pixels})
					arr = arr[3:]
				}
			}
			return icons.SaveAsPngToSessionIconDir(res)
		}
	}
	return ""
}
