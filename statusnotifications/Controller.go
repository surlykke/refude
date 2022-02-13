// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"errors"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/icons"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
	"github.com/surlykke/RefudeServices/lib/dbusutils"
	"github.com/surlykke/RefudeServices/lib/log"
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
 * serviceName Can be a name of service or a path of object
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
			log.Info("Ignoring signal", signal.Name, "from", signal.Sender, signal.Path)
		}
	}
}

func checkItemStatus(sender string) {
	for _, res := range Items.GetAll() {
		var item = res.(*Item)
		if item.sender == sender {
			if _, ok := dbuscall.GetSingleProp(conn, item.sender, item.path, ITEM_INTERFACE, "Status"); !ok {
				events <- Event{"ItemRemoved", item.sender, item.path}
			}
		}
	}
}

// ----------------------------------------------------------------------------------------------------

func getOnTheBus() {
	var err error

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
				log.Info("Got UnregisterStatusNotifierItem:", s, ",", sender)
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

type Event struct {
	eventName string // New item: "ItemCreated", otherwise name of relevant dbus signal
	sender    string
	path      dbus.ObjectPath
}

var events = make(chan Event)

func updateWatcherProperties() {
	ids := make([]string, 0, 20)
	for _, res := range Items.GetAll() {
		var item = res.(*Item)
		ids = append(ids, item.sender+":"+string(item.path))
	}
	watcherProperties.Set(WATCHER_INTERFACE, "RegisteredStatusItems", dbus.MakeVariant(ids))
}

func buildItem(sender string, path dbus.ObjectPath) *Item {
	var item = Item{sender: sender, path: path}
	var props = dbuscall.GetAllProps(conn, item.sender, item.path, ITEM_INTERFACE)
	item.Id = getStringOr(props["Id"])
	item.Category = getStringOr(props["Category"])
	item.MenuPath = getDbusPath(props["Menu"])
	item.Title = getStringOr(props["Title"])
	item.Status = getStringOr(props["Status"])
	item.ToolTip = getStringOr(props["ToolTip"])

	if item.IconThemePath = getStringOr(props["IconThemePath"]); item.IconThemePath != "" {
		icons.AddBasedir(item.IconThemePath)
	}

	if item.UseIconPixmap = getStringOr(props["IconName"]) == ""; item.UseIconPixmap {
		item.IconName = collectPixMap(props["IconPixmap"])
	} else {
		item.IconName = getStringOr(props["IconName"])
	}

	if item.UseAttentionIconPixmap = getStringOr(props["AttentionIconName"]) == ""; item.UseAttentionIconPixmap {
		item.AttentionIconName = collectPixMap(props["AttentionIconPixmap"])
	} else {
		item.AttentionIconName = getStringOr(props["AttentionIconName"])
	}

	item.UseOverlayIconPixmap = getStringOr(props["OverlayIconName"]) == "" // TODO

	return &item
}

func getProp(sender string, path dbus.ObjectPath, propname string) (dbus.Variant, bool) {
	return dbuscall.GetSingleProp(conn, sender, path, ITEM_INTERFACE, propname)
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
