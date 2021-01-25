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
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/icons"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
	dbuscall "github.com/surlykke/RefudeServices/lib/dbusutils"
	"github.com/surlykke/RefudeServices/lib/respond"
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
	var item = Item{self: itemSelf(sender, path), sender: sender, itemPath: path}
	var props = dbuscall.GetAllProps(conn, item.sender, item.itemPath, ITEM_INTERFACE)
	item.Id = getStringOr(props["Id"])
	item.Category = getStringOr(props["Category"])
	if item.menuPath = getDbusPath(props["Menu"]); item.menuPath != "" {
		item.Menu = item.self + "/menu"
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

	item.Links = respond.Links{item.itemLink(respond.Self)}
	if item.menuPath != "" {
		item.Links = append(item.Links, item.menuLink(respond.Related))
	}
	return &item
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
	if len(item.Links) > 0 && item.Links[0].Rel == respond.Self {
		item.Links[0].Icon = icons.IconUrl(item.IconName)
	}
}

func updateIconThemePath(item *Item) {
	if v, ok := getProp(item, "IconThemePath"); ok {
		item.iconThemePath = getStringOr(v)
		//icons.AddBaseDir(item.iconThemePath)
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
