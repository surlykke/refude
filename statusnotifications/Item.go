// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
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
	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/slice"
)

type Item struct {
	sender                  string
	itemPath                dbus.ObjectPath
	self                    string
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
	MenuPath                dbus.ObjectPath
	IconThemePath           string
	UseIconPixmap           bool
	UseAttentionIconPixmap  bool
	UseOverlayIconPixmap    bool
}

func (item *Item) Links() []resource.Link {
	var links []resource.Link = []resource.Link{resource.MakeLink(item.self, item.Title, item.IconName, relation.Self)}
	if item.MenuPath != "" {
		links = append(links, resource.MakeLink(item.self+"/menu", item.Title, "", relation.Menu))
	}
	return links
}

func (item *Item) RefudeType() string {
	return "item"
}

func (item *Item) buildMenu() *Menu {
	if item.MenuPath == "" {
		return nil
	} else {
		if entries, err := menuEntries(item.sender, item.MenuPath); err != nil {
			return nil
		} else {
			return &Menu{
				Entries:  entries,
				sender:   item.sender,
				menuPath: item.MenuPath,
				self:     item.self + "/menu",
			}
		}

	}
}

func itemSelf(sender string, path dbus.ObjectPath) string {
	return fmt.Sprintf("/item/%s", strings.ReplaceAll(sender+string(path), "/", "-"))
}

type ItemMap map[string]*Item

type MenuEntry struct {
	Id          string
	Type        string
	Label       string
	Enabled     bool
	Visible     bool
	IconName    string
	Shortcuts   [][]string `json:",omitempty"`
	ToggleType  string     `json:",omitempty"`
	ToggleState int32
	SubEntries  []MenuEntry `jsoControllern:",omitempty"`
}

type Menu struct {
	Entries  []MenuEntry
	sender   string
	menuPath dbus.ObjectPath
	self     string
}

func (menu *Menu) Links() []resource.Link {
	return []resource.Link{resource.MakeLink(menu.self, "", "", relation.Self)}
}

func (menu *Menu) RefudeType() string {
	return "menu"
}

func menuEntries(sender string, menuPath dbus.ObjectPath) ([]MenuEntry, error) {
	obj := conn.Object(sender, menuPath)
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
		return []MenuEntry{menu}, nil
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

	menuItem.Type = getStringOr(m["type"])
	if menuItem.Type == "" {
		menuItem.Type = "standard"
	}
	if !slice.Among(menuItem.Type, "standard", "separator") {
		return MenuEntry{}, errors.New("Illegal menuitem type: " + menuItem.Type)
	}
	menuItem.Label = getStringOr(m["label"])
	menuItem.Enabled = getBoolOr(m["enabled"], true)
	menuItem.Visible = getBoolOr(m["visible"], true)
	if menuItem.IconName = getStringOr(m["icon-name"]); menuItem.IconName == "" {
		// TODO: Look for pixmap
	}
	if menuItem.ToggleType = getStringOr(m["toggle-type"]); !slice.Among(menuItem.ToggleType, "checkmark", "radio", "") {
		return MenuEntry{}, errors.New("Illegal toggle-type: " + menuItem.ToggleType)
	}

	menuItem.ToggleState = getInt32Or(m["toggle-state"], -1)
	if childrenDisplay := getStringOr(m["children-display"]); childrenDisplay == "submenu" {
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

func collectPixMap(variant dbus.Variant) string {
	if arrs, ok := variant.Value().([][]interface{}); ok {
		var images = []image.ARGBImage{}
		for _, arr := range arrs {
			for len(arr) > 2 {
				width := uint32(arr[0].(int32))
				height := uint32(arr[1].(int32))
				pixels := arr[2].([]byte)
				images = append(images, image.ARGBImage{Width: width, Height: height, Pixels: pixels})
				arr = arr[3:]
			}
		}
		var argbIcon = image.ARGBIcon{Images: images}
		return icons.AddARGBIcon(argbIcon)
	}
	return ""
}

func (m *Menu) DoPost(w http.ResponseWriter, r *http.Request) {
	var menuId = requests.GetSingleQueryParameter(r, "id", "")
	for _, entry := range m.Entries {
		if menuId == entry.Id {
			idAsInt, _ := strconv.Atoi(menuId)
			data := dbus.MakeVariant("")
			time := uint32(time.Now().Unix())
			dbusObj := conn.Object(m.sender, m.menuPath)
			if err := dbusObj.Call("com.canonical.dbusmenu.Event", dbus.Flags(0), idAsInt, "clicked", data, time).Err; err != nil {
				respond.ServerError(w, err)
			} else {
				respond.Accepted(w)
			}
			return
		}
	}
	respond.NotFound(w)
}
