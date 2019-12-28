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
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"

	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"github.com/godbus/dbus/prop"
	dbuscall "github.com/surlykke/RefudeServices/lib/dbusutils"
	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/slice"
)

const WATCHER_SERVICE = "org.kde.StatusNotifierWatcher"
const WATCHER_PATH = "/StatusNotifierWatcher"
const WATCHER_INTERFACE = WATCHER_SERVICE
const HOST_SERVICE = "org.kde.StatusNotifierHost"
const ITEM_PATH = "/StatusNotifierItem"
const ITEM_INTERFACE = "org.kde.StatusNotifierItem"
const MENU_INTERFACE = "com.canonical.dbusmenu"

var conn *dbus.Conn
var watcherProperties *prop.Properties

func senderAndPath(serviceName string, sender dbus.Sender) (string, dbus.ObjectPath) {
	if regexp.MustCompile("^(/\\w+)+$").MatchString(serviceName) {
		return string(sender), dbus.ObjectPath(serviceName)
	} else {
		return string(sender), dbus.ObjectPath(ITEM_PATH)
	}
}

/**
 * serviceId Can be a name of service or a path of object
 */
func addItem(serviceName string, sender dbus.Sender) *dbus.Error {
	var s, p = senderAndPath(serviceName, sender)
	events <- Event{"ItemCreated", s, p}
	return nil
}

func monitorSignals() {
	var dbusSignals = make(chan *dbus.Signal, 50)
	conn.Signal(dbusSignals)
	addMatch := "org.freedesktop.DBus.AddMatch"
	conn.BusObject().Call(addMatch, 0, "type='signal', interface='org.kde.StatusNotifierItem'")
	conn.BusObject().Call(addMatch, 0, "sender=org.freedesktop.DBus,path=/org/freedesktop/DBus,interface=org.freedesktop.DBus,member=NameOwnerChanged,type=signal")
	for signal := range dbusSignals {
		if signal.Name == "org.freedesktop.DBus.NameOwnerChanged" {
			if len(signal.Body) >= 3 {
				if oldOwner, ok := signal.Body[1].(string); ok {
					checkItemStatus(oldOwner)
				}
			}
		} else if strings.HasPrefix(signal.Name, "org.kde.StatusNotifierItem.New") {
			events <- Event{signal.Name, signal.Sender, signal.Path}
		} else {
			fmt.Println("Ignoring signal", signal.Name, "from", signal.Sender, signal.Path)
		}
	}
}

func checkItemStatus(sender string) {
	for _, item := range items {
		if item.sender == sender {
			if _, ok := dbuscall.GetSingleProp(conn, item.sender, item.itemPath, ITEM_INTERFACE, "Status"); !ok {
				events <- Event{"ItemRemoved", item.sender, item.itemPath}
			}
		}
	}
}

// ----------------------------------------------------------------------------------------------------

func getOnTheBus() {
	var err error

	// Get on the bus
	conn, err = dbus.SessionBus()
	if err != nil {
		panic(err)
	} else if reply, err := conn.RequestName(WATCHER_SERVICE, dbus.NameFlagDoNotQueue); err != nil {
		panic(err)
	} else if reply != dbus.RequestNameReplyPrimaryOwner {
		panic(errors.New(WATCHER_SERVICE + " taken"))
	} else if reply, err = conn.RequestName(HOST_SERVICE, dbus.NameFlagDoNotQueue); err != nil {
		panic(err)
	} else if reply != dbus.RequestNameReplyPrimaryOwner {
		panic(errors.New(HOST_SERVICE + " taken"))
	}

	// Put StatusNotifierWatcher object up
	_ = conn.ExportMethodTable(
		map[string]interface{}{
			"RegisterStatusNotifierItem": addItem,
			"UnregisterStatusNotifierItem": func(s string, sender dbus.Sender) {
				fmt.Println("Got UnregisterStatusNotifierItem:", s, ",", sender)
			}, // We dont care, see monitorItem
		},
		WATCHER_PATH,
		WATCHER_INTERFACE,
	)

	// Add Introspectable interface
	_ = conn.Export(introspect.Introspectable(INTROSPECT_XML), WATCHER_PATH, dbuscall.INTROSPECT_INTERFACE)

	// Add properties interface
	watcherProperties = prop.New(
		conn,
		WATCHER_PATH,
		map[string]map[string]*prop.Prop{
			WATCHER_INTERFACE: {
				"IsStatusNotifierHostRegistered": {Value: true, Writable: false, Emit: prop.EmitTrue, Callback: nil},
				"ProtocolVersion":                {Value: 0, Writable: false, Emit: prop.EmitTrue, Callback: nil},
				"RegisteredStatusItems":          {Value: []string{}, Writable: false, Emit: prop.EmitTrue, Callback: nil},
			},
		},
	)
}

type EventType int

const (
	ItemCreated EventType = iota
	TitleUpdated
	IconUpdated
	AttentionIconUpdated
	OverlayIconUpdated
	ToolTipUpdated
	StatusUpdated
	ItemRemoved
)

type Event struct {
	eventName string // New item: "ItemCreated", otherwise name of relevant dbus signal
	sender    string
	path      dbus.ObjectPath
}

var events = make(chan Event)

func updateWatcherProperties() {
	ids := make([]string, 0, 20)
	for _, item := range items {
		ids = append(ids, item.sender+":"+string(item.itemPath))
	}
	watcherProperties.Set(WATCHER_INTERFACE, "RegisteredStatusItems", dbus.MakeVariant(ids))
}

func buildItem(sender string, path dbus.ObjectPath) *Item {
	fmt.Println("MakeItem")
	var item = MakeItem(sender, path)
	var props = dbuscall.GetAllProps(conn, item.sender, item.itemPath, ITEM_INTERFACE)
	item.Id = getStringOr(props["Id"])
	item.Category = getStringOr(props["Category"])
	if menuPath := getDbusPath(props["Menu"]); menuPath != "" {
		item.menu = MakeMenuResource(item.sender, menuPath)
		item.Menu = item.menu.self
	}
	item.Title = getStringOr(props["Title"])
	item.Status = getStringOr(props["Status"])
	item.ToolTip = getStringOr(props["ToolTip"])

	if iconThemePath := getStringOr(props["IconThemePath"]); iconThemePath != "" {
		icons.AddBaseDir(iconThemePath)
	}

	if item.useIconPixmap = getStringOr(props["IconName"]) == ""; item.useIconPixmap {
		item.IconName = collectPixMap(props["IconPixmap"])
	} else {
		item.IconName = getStringOr(props["IconName"])
	}

	if item.useAttentionIconPixmap = getStringOr(props["AttentionIconName"]) == ""; item.useAttentionIconPixmap {
		item.AttentionIconName = collectPixMap(props["AttentionIconPixmap"])
	} else {
		item.AttentionIconName = getStringOr(props["AttentionIconName"])
	}

	item.useOverlayIconPixmap = getStringOr(props["OverlayIconName"]) == "" // TODO

	return item
}

func updateTitle(item *Item) {
	if v, ok := getProp(item, "Title"); ok {
		item.Title = getStringOr(v)
	}
}

func updateToolTip(item *Item) {
	if v, ok := getProp(item, "ToolTip"); ok {
		item.ToolTip = getStringOr(v)
	}
}

func updateStatus(item *Item) {
	if v, ok := getProp(item, "Status"); ok {
		item.Status = getStringOr(v)
	}
}

func updateIcon(item *Item) {
	if item.useIconPixmap {
		if v, ok := getProp(item, "IconPixmap"); ok {
			item.IconName = collectPixMap(v)
		}
	} else {
		if v, ok := getProp(item, "IconName"); ok {
			item.IconName = getStringOr(v)
		}
	}
}

func updateAttentionIcon(item *Item) {
	if item.useAttentionIconPixmap {
		if v, ok := getProp(item, "AttentionIconPixmap"); ok {
			item.AttentionIconName = collectPixMap(v)
		}
	} else {
		if v, ok := getProp(item, "AttentionIconName"); ok {
			item.AttentionIconName = getStringOr(v)
		}
	}
}

func updateOverlayIcon(item *Item) {
	// TODO
}

func getProp(item *Item, propname string) (dbus.Variant, bool) {
	return dbuscall.GetSingleProp(conn, item.sender, item.itemPath, ITEM_INTERFACE, propname)
}

func getStringOr(v dbus.Variant) string {
	if res, ok := v.Value().(string); ok {
		return res
	}

	return ""
}

func getBoolOr(variant dbus.Variant, fallback bool) bool {
	if res, ok := variant.Value().(bool); ok {
		return res
	}

	return fallback
}

func getInt32Or(variant dbus.Variant, fallback int32) int32 {
	if res, ok := variant.Value().(int32); ok {
		return res
	}
	return fallback
}

func getDbusPath(variant dbus.Variant) dbus.ObjectPath {
	if res, ok := variant.Value().(dbus.ObjectPath); ok {
		return res
	}
	return ""
}

func parseMenu(value []interface{}) (MenuItem, error) {
	var menuItem = MenuItem{}
	var id int32
	var ok bool
	var m map[string]dbus.Variant
	var s []dbus.Variant

	if len(value) < 3 {
		return MenuItem{}, errors.New("Wrong length")
	} else if id, ok = value[0].(int32); !ok {
		return MenuItem{}, errors.New("Expected int32, got: " + reflect.TypeOf(value[0]).String())
	} else if m, ok = value[1].(map[string]dbus.Variant); !ok {
		return MenuItem{}, errors.New("Excpected dbus.Variant, got: " + reflect.TypeOf(value[1]).String())
	} else if s, ok = value[2].([]dbus.Variant); !ok {
		return MenuItem{}, errors.New("expected []dbus.Variant, got: " + reflect.TypeOf(value[2]).String())
	}

	menuItem.Id = fmt.Sprintf("%d", id)

	menuItem.Type = getStringOr(m["type"])
	if menuItem.Type == "" {
		menuItem.Type = "standard"
	}
	if !slice.Among(menuItem.Type, "standard", "separator") {
		return MenuItem{}, errors.New("Illegal menuitem type: " + menuItem.Type)
	}
	menuItem.Label = getStringOr(m["label"])
	menuItem.Enabled = getBoolOr(m["enabled"], true)
	menuItem.Visible = getBoolOr(m["visible"], true)
	if menuItem.IconName = getStringOr(m["icon-name"]); menuItem.IconName == "" {
		// FIXME: Look for pixmap
	}
	if menuItem.ToggleType = getStringOr(m["toggle-type"]); !slice.Among(menuItem.ToggleType, "checkmark", "radio", "") {
		return MenuItem{}, errors.New("Illegal toggle-type: " + menuItem.ToggleType)
	}

	menuItem.ToggleState = getInt32Or(m["toggle-state"], -1)
	if childrenDisplay := getStringOr(m["children-display"]); childrenDisplay == "submenu" {
		for _, variant := range s {
			if interfaces, ok := variant.Value().([]interface{}); !ok {
				return MenuItem{}, errors.New("Submenu item not of type []interface")
			} else {
				if submenuItem, err := parseMenu(interfaces); err != nil {
					return MenuItem{}, err
				} else {
					menuItem.SubEntries = append(menuItem.SubEntries, submenuItem)
				}
			}
		}
	} else if childrenDisplay != "" {
		log.Println("warning: ignoring unknown children-display type:", childrenDisplay)
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
