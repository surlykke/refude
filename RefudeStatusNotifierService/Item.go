// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/surlykke/RefudeServices/lib/service"
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
)


type Item struct {
	Id string
	Category string
	Status string
	IconName string
	IconAccessibleDesc string
	AttentionIconName string
	AttentionAccessibleDesc string
	Title string
	Menu []MenuItem `json:",omitempty"`

	iconThemePath string
	menuObject dbus.BusObject
	menu dbus.BusObject
	dbusObj dbus.BusObject
	path string
}

type MenuItem struct {
	Id int32
	Type string
	Label string
	Enabled bool
	Visible bool
	IconName string
	Shortcuts [][]string
	ToggleType string
	TogleState int
	ChildrenDisplay string
	SubMenus []MenuItem `json:",omitempty"`
}




func (item *Item) GET(w http.ResponseWriter, r *http.Request) {
	resource.JsonGET(item, w)
}

func (item *Item) POST(w http.ResponseWriter, r *http.Request) {
	method := resource.GetSingleQueryParameter(r, "method", "Activate")
	x, errX := strconv.Atoi(resource.GetSingleQueryParameter(r, "x", "0"))
	y, errY := strconv.Atoi(resource.GetSingleQueryParameter(r, "y", "0"))
	if (method != "Activate" && method != "SecondaryActivate" && method != "ContextMenu") ||
		errX != nil ||
		errY != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	item.dbusObj.Call("org.kde.StatusNotifierItem."+method, dbus.Flags(0), x, y)
	w.WriteHeader(http.StatusAccepted)
}

func (item *Item) copy() *Item {
	var tmp = *item
	return &tmp
}

func (item *Item) getStringProp(propName string) string {
	var method = PROPERTIES_INTERFACE + ".Get"
	if call := item.dbusObj.Call(method, dbus.Flags(0), ITEM_INTERFACE, propName);  call.Err != nil {
		log.Println(call.Err)
		return ""
	} else if value, ok := call.Body[0].(dbus.Variant).Value().(string); !ok {
		log.Printf("Property '%s' not of type string", propName)
		return ""
	} else {
		return value
	}
}

func (item *Item) getPathProp(propName string) dbus.ObjectPath {
	var method = PROPERTIES_INTERFACE + ".Get"
	if call := item.dbusObj.Call(method, dbus.Flags(0), ITEM_INTERFACE, propName);  call.Err != nil {
		log.Println(call.Err)
		return ""
	} else if value, ok := call.Body[0].(dbus.Variant).Value().(dbus.ObjectPath); !ok {
		log.Printf("Property '%s' not of type dbus.ObjectPath", propName)
		return ""
	} else {
		return value
	}
}

func (item *Item) getMenu() {
	var method = MENU_INTERFACE + ".GetLayout"
	if call := item.menuObject.Call(method, dbus.Flags(0), 0, -1, []string{}); call.Err != nil {
		log.Println(call.Err)
	} else if len(call.Body) < 2 {
		log.Println("Retrieved", len(call.Body), "arguments, expected 2")
	} else if _, ok := call.Body[0].(uint32); !ok {
		log.Println("Expected uint32 as first return argument, got:", reflect.TypeOf(call.Body[0]))
	} else if interfaces, ok := call.Body[1].([]interface{}); !ok {
		log.Println("Expected []interface{} as second return argument, got:", reflect.TypeOf(call.Body[1]))
	} else if menu, err := parseMenu(interfaces); err != nil {
		log.Println("Error retrieving menu", err)
	} else if len(menu.SubMenus) > 0 {
		item.Menu = menu.SubMenus
	} else {
		item.Menu = []MenuItem{menu}
	}
}


func getString(variantMap map[string]dbus.Variant, key string, fallback string) string {
	if v, ok := variantMap[key]; ok {
		if res, ok := v.Value().(string); ok {
			return res
		}
	}

	return fallback
}

func getBoolean(variantMap map[string]dbus.Variant, key string, fallback bool) bool {
	if v, ok := variantMap[key]; ok {
		if res, ok := v.Value().(bool); ok {
			return res
		}
	}

	return fallback
}

func getInt(variantMap map[string]dbus.Variant, key string, fallback int) int {
	if v, ok := variantMap[key]; ok {
		if res, ok := v.Value().(int); ok {
			return res
		}
	}

	return fallback
}

func parseMenu(value []interface{}) (MenuItem, error) {
	fmt.Println("")
	var menuItem = MenuItem{}
	var ok bool
	var m map[string]dbus.Variant
	var s []dbus.Variant

	if len(value) < 3 {
		return MenuItem{}, errors.New("Wrong length")
	} else if menuItem.Id, ok = value[0].(int32); !ok {
		return MenuItem{}, errors.New("Expected int32, got: " + reflect.TypeOf(value[0]).String())
	} else if m,ok = value[1].(map[string]dbus.Variant); !ok {
		return MenuItem{}, errors.New("Excpected dbus.Variant, got: " + reflect.TypeOf(value[1]).String())
	} else if s,ok = value[2].([]dbus.Variant); !ok {
		return MenuItem{}, errors.New("expected []dbus.Variant, got: " + reflect.TypeOf(value[2]).String())
	}

	if menuItem.Type = getString(m, "type", "standard"); !utils.Among(menuItem.Type, "standard", "separator") {
		return MenuItem{}, errors.New("Illegal menuitem type: " + menuItem.Type)
	}
	menuItem.Label = getString(m, "label", "")
	menuItem.Enabled = getBoolean(m, "enabled", true)
	menuItem.Visible = getBoolean(m, "visible", true)
	if menuItem.IconName = getString(m, "icon-name", ""); menuItem.IconName == "" {
		// FIXME: Look for pixmap
	}
	if menuItem.ToggleType = getString(m, "toggle-type", ""); !utils.Among(menuItem.ToggleType, "checkmark", "radio", "") {
		return MenuItem{}, errors.New("Illegal toggle-type: " + menuItem.ToggleType)
	}

	menuItem.TogleState = getInt(m, "toggle-state", -1)
	if childrenDisplay := getString(m, "children-display", ""); childrenDisplay == "submenu" {
		for _,variant := range s {
			if interfaces, ok := variant.Value().([]interface{}); !ok {
				return MenuItem{}, errors.New("Submenu item not of type []interface")
			} else {
				if submenuItem, err := parseMenu(interfaces); err != nil {
					return MenuItem{}, err
				} else {
					menuItem.SubMenus = append(menuItem.SubMenus, submenuItem)
				}
			}
		}
	} else if childrenDisplay != "" {
		log.Println("warning: ignoring unknown children-display type:", childrenDisplay)
	}

	return menuItem, nil
}


type SenderAndPath struct {
	sender  string
	objPath dbus.ObjectPath
}

func MakeItem(sp SenderAndPath) *Item {
	var item = &Item{}
	item.dbusObj = conn.Object(sp.sender, sp.objPath)
	item.Id = item.getStringProp("Id")
	item.Category = item.getStringProp("Category")
	item.Status = item.getStringProp("Status")
	item.iconThemePath = item.getStringProp("IconThemePath");
	item.IconName = item.getStringProp("IconName")
	if item.IconName == "" {
		item.IconName = item.getPixMap("IconPixmap")
	} else if item.iconThemePath != ""{
		icons.CopyIcons(item.IconName, item.iconThemePath)
	}
	item.IconAccessibleDesc = item.getStringProp("IconAccessibleDesc")
	item.AttentionIconName = item.getStringProp("AttentionIconName")
	if item.AttentionIconName == ""	{
		item.AttentionIconName = item.getPixMap("AttentionIconPixmap")
	} else if item.iconThemePath != "" {
		icons.CopyIcons(item.AttentionIconName, item.iconThemePath)
	}
	item.AttentionAccessibleDesc = item.getStringProp("AttentionAccessibleDesc")
	item.Title = item.getStringProp("Title")
	var menuPath = item.getPathProp("Menu")
	if menuPath != "" {
		item.menuObject = conn.Object(sp.sender, menuPath)
		item.getMenu()
	}

	return item
}

func (item *Item) Run(path string, signals chan *dbus.Signal) {
	// TODO Menu
	service.Map(path, item.copy())
	defer service.Unmap(path)

	for signal := range signals {
		fmt.Print(path, " signal: ", signal.Name, signal.Body, "\n")
		switch signal.Name {
		case "org.kde.StatusNotifierItem.NewIcon":
			item.IconName = item.getStringProp("IconName")
			item.IconAccessibleDesc = item.getStringProp("IconAccessibleDesc")
			if item.IconName == "" {
				item.IconName = item.getPixMap("IconPixmap")
			} else if item.iconThemePath != ""{
				icons.CopyIcons(item.IconName, item.iconThemePath)
			}
			service.Map(path, item.copy())
		case "org.kde.StatusNotifierItem.NewAttentionIcon":
			item.AttentionIconName = item.getStringProp("AttentionIconName")
			item.AttentionAccessibleDesc = item.getStringProp("AttentionAccessibleDesc")
			if item.AttentionIconName == ""	{
				item.AttentionIconName = item.getPixMap("AttentionIconName")
			} else if item.iconThemePath != "" {
				icons.CopyIcons(item.AttentionIconName, item.iconThemePath)
			}
			service.Map(path, item.copy())
		case "org.kde.StatusNotifierItem.NewStatus":
			if tmp, ok := signal.Body[0].(string); ok {
				item.Status = tmp;
			} else {
				log.Println("NewStatus signal: ", signal.Body[0], ", not a string")
			}
			service.Map(path, item.copy())
		case "org.kde.StatusNotifierItem.NewIconThemePath":
			item.iconThemePath,_ = signal.Body[0].(string)
			if item.iconThemePath != "" {
				if item.IconName != "" {
					icons.CopyIcons(item.IconName, item.iconThemePath)
				}
				if item.AttentionIconName != "" {
					icons.CopyIcons(item.AttentionIconName, item.iconThemePath)
				}
			}
		}
	}
}

func (item *Item) getPixMap(propName string) string {
	if call := item.dbusObj.Call(PROPERTIES_INTERFACE+".Get", dbus.Flags(0), ITEM_INTERFACE, propName); call.Err != nil {
		return ""
	} else {
		value := call.Body[0].(dbus.Variant).Value()
		dbusValue, ok := value.([][]interface{})
		if !ok {
			log.Println("Expected", propName, "to be of type [][]interface{}, but found", reflect.TypeOf(value))
			return ""
		}

		return collectPixMap(dbusValue)
	}

}

func collectPixMap(dbusValue [][]interface{}) string {
	res := make(icons.Icon, 0)
	for _, arr := range (dbusValue) {
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
