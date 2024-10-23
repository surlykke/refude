// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package statusnotifications

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/icon"
	"github.com/surlykke/RefudeServices/icons"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
	dbuscall "github.com/surlykke/RefudeServices/lib/dbusutils"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/path"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/resource"
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
	events <- event{"ItemCreated", s, p}
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
			events <- event{signal.Name, signal.Sender, signal.Path}
		} else {
			log.Info("Ignoring signal", signal.Name, "from", signal.Sender, signal.Path)
		}
	}
}

func checkItemStatus(sender string) {
	for _, item := range repo.GetList[*Item]("/item/") {
		if item.DbusSender == sender {
			if _, ok := dbuscall.GetSingleProp(conn, item.DbusSender, item.DbusPath, ITEM_INTERFACE, "Status"); !ok {
				events <- event{"ItemRemoved", item.DbusSender, item.DbusPath}
			}
		}
	}
}

// ----------------------------------------------------------------------------------------------------

func getOnTheBus() {
	var err error

	defer func() {
		if err := recover(); err != nil {
			log.Warn(err, "- hence statusnotifications not running")
		}
	}()

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
	watcherProperties, _ = prop.Export(
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

type event struct {
	name       string // New item: "ItemCreated", otherwise name of relevant dbus signal
	dbusSender string
	dbusPath   dbus.ObjectPath
}

var events = make(chan event)

func buildItem(itemPath path.Path, dbusSender string, dbusPath dbus.ObjectPath) *Item {
	var item = Item{
		ResourceData: *resource.MakeBase(itemPath, "", "", "", mediatype.Trayitem),
		DbusSender:   dbusSender,
		DbusPath:     dbusPath,
	}

	if err := conn.BusObject().Call("org.freedesktop.DBus.GetConnectionUnixProcessID", 0, dbusSender).Store(&item.SenderPid); err != nil {
		log.Warn("get processid err:", err)
	} else if bytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", item.SenderPid)); err != nil {
		log.Warn("Error reading proc cmdline for", item.SenderPid)
	} else {
		item.SenderApp = extractCommandName(bytes)
	}

	var props = dbuscall.GetAllProps(conn, item.DbusSender, item.DbusPath, ITEM_INTERFACE)
	item.ItemId = getString(props["Id"])
	item.Category = getString(props["Category"])
	item.MenuDbusPath = getDbusPath(props["Menu"])
	if item.MenuDbusPath != "" {
		item.MenuPath = path.Of("/menu/", dbusSender, item.MenuDbusPath)
	}
	if item.IconThemePath = getString(props["IconThemePath"]); item.IconThemePath != "" {
		icons.AddBasedir(item.IconThemePath)
	}

	RetrieveTitle(&item)
	RetrieveStatus(&item)
	RetrieveToolTip(&item)

	RetrieveIcon(&item)
	RetrieveAttentionIcon(&item)
	RetrieveOverlayIcon(&item)
	return &item
}

var cmdLineArg = regexp.MustCompile(` -\S*`)
var repeatedSpaces = regexp.MustCompile(`  +`)

func extractCommandName(bytes []byte) string {
	var cmdline = string(bytes)

	// The kernel uses \0 as separator
	cmdline = strings.ReplaceAll(cmdline, "\u0000", " ")

	// Remove upto and including last slash
	// so eg:
	//    - "nm-applet" is unchanged
	//    - "/opt/google/chrome/chrome --arg1 --arg2" becomes "chrome --arg1 --arg2"
	//    - "/usr/bin/python3 /usr/bin/blueman-tray"  becomes "blueman-tray"
	//
	// Something like /path/to/command /path/to/file-arg will not work well.
	//
	cmdline = cmdline[strings.LastIndex(cmdline, "/")+1:]

	// Remove trailing args, so eg. "chrome --arg1 --arg2" becomes "chrome"
	if firstSpace := strings.Index(cmdline, " "); firstSpace > -1 {
		cmdline = cmdline[:firstSpace]
	}

	return cmdline
}

//case "NewTitle", "NewIcon", "NewAttentionIcon", "NewOverlayIcon", "NewToolTip","NewStatus":

func RetrieveTitle(item *Item) {
	if prop, ok := getProp(item.DbusSender, item.DbusPath, "Title"); ok {
		item.Title = getString(prop)
	}
}

func RetrieveIcon(item *Item) {
	if prop, ok := getProp(item.DbusSender, item.DbusPath, "IconName"); ok {
		item.Icon = icon.Name(getString(prop))
	} else if prop, ok = getProp(item.DbusSender, item.DbusPath, "IconPixmap"); ok {
		item.Icon = collectPixMap(prop)
	}
}

func RetrieveAttentionIcon(item *Item) {
	if prop, ok := getProp(item.DbusSender, item.DbusPath, "AttentionIconName"); ok {
		item.AttentionIconName = icon.Name(getString(prop))
	} else if prop, ok = getProp(item.DbusSender, item.DbusPath, "AttentionIconPixmap"); ok {
		item.AttentionIconName = collectPixMap(prop)
	}
}

func RetrieveOverlayIcon(item *Item) {
	if prop, ok := getProp(item.DbusSender, item.DbusPath, "OverlayIconName"); ok {
		item.OverlayIconName = icon.Name(getString(prop))
	} else if prop, ok = getProp(item.DbusSender, item.DbusPath, "OverlayIconPixmap"); ok {
		item.OverlayIconName = collectPixMap(prop)
	}
}

func RetrieveToolTip(item *Item) {
	if prop, ok := getProp(item.DbusSender, item.DbusPath, "ToolTip"); ok {
		item.ToolTip = getString(prop)
	}
}

func RetrieveStatus(item *Item) {
	if prop, ok := getProp(item.DbusSender, item.DbusPath, "Status"); ok {
		item.Status = getString(prop)
	}
}

func getProp(sender string, path dbus.ObjectPath, propname string) (dbus.Variant, bool) {
	return dbuscall.GetSingleProp(conn, sender, path, ITEM_INTERFACE, propname)
}

func getString(v dbus.Variant) string {
	if res, ok := v.Value().(string); ok {
		return res
	}

	return ""
}

func getBool(variant dbus.Variant, fallback bool) bool {
	if res, ok := variant.Value().(bool); ok {
		return res
	}

	return fallback
}

func getInt32(variant dbus.Variant, fallback int32) int32 {
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

func getBytes(variant dbus.Variant) []byte {
	if bytes, ok := variant.Value().([]byte); ok {
		return bytes
	} else {
		return nil
	}
}
