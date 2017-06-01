package main

import (
	"github.com/godbus/dbus"
	"errors"
	"fmt"
	"github.com/godbus/dbus/prop"
	"sync"
	"github.com/godbus/dbus/introspect"
	"github.com/surlykke/RefudeServices/lib/stringlist"
	"strings"
)

// Takes care of the dbus-side of things

const WATCHER_SERVICE = "org.kde.StatusNotifierWatcher"
const WATCHER_PATH = "/StatusNotifierWatcher"
const WATCHER_INTERFACE = WATCHER_SERVICE
const HOST_SERVICE = "org.kde.StatusNotifierHost"
const ITEM_PATH = "/StatusNotifierItem"
const ITEM_INTERFACE = "org.kde.StatusNotifierItem"
const INTROSPECT_INTERFACE = "org.freedesktop.DBus.Introspectable"
const INTROSPECT_XML =
`<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
        "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd">
<node>
    <interface name="org.kde.StatusNotifierWatcher">
        <property name="IsStatusNotifierHostRegistered" type="b" access="read"/>
        <property name="ProtocolVersion" type="i" access="read"/>
        <property name="RegisteredStatusNotifierItems" type="as" access="read"/>
        <signal name="StatusNotifierItemRegistered">
            <arg name="service" type="s" direction="out"/>
        </signal>
        <signal name="StatusNotifierItemUnregistered">
            <arg name="service" type="s" direction="out"/>
        </signal>
        <signal name="StatusNotifierHostRegistered">
        </signal>
        <method name="RegisterStatusNotifierItem">
            <arg name="serviceOrPath" type="s" direction="in"/>
        </method>
        <method name="RegisterStatusNotifierHost">
            <arg name="service" type="s" direction="in"/>
        </method>
    </interface>
    <interface name="org.freedesktop.DBus.Properties">
        <method name="Get">
            <arg name="interface_name" type="s" direction="in"/>
            <arg name="property_name" type="s" direction="in"/>
            <arg name="value" type="v" direction="out"/>
        </method>
        <method name="Set">
            <arg name="interface_name" type="s" direction="in"/>
            <arg name="property_name" type="s" direction="in"/>
            <arg name="value" type="v" direction="in"/>
        </method>
        <method name="GetAll">
            <arg name="interface_name" type="s" direction="in"/>
            <arg name="values" type="a{sv}" direction="out"/>
            <annotation name="org.qtproject.QtDBus.QtTypeName.Out0" value="QVariantMap"/>
        </method>
        <signal name="PropertiesChanged">
            <arg name="interface_name" type="s" direction="out"/>
            <arg name="changed_properties" type="a{sv}" direction="out"/>
            <arg name="invalidated_properties" type="as" direction="out"/>
        </signal>
    </interface>
    <interface name="org.freedesktop.DBus.Introspectable">
        <method name="Introspect">
            <arg name="xml_data" type="s" direction="out"/>
        </method>
    </interface>
    <interface name="org.freedesktop.DBus.Peer">
        <method name="Ping"/>
        <method name="GetMachineId">
            <arg name="machine_uuid" type="s" direction="out"/>
        </method>
    </interface>
</node>`

type Props map[string]dbus.Variant
type PropChangeChannel chan Props
type ChannelMap map[string]PropChangeChannel

var	conn *dbus.Conn
var dbusSignals = make(chan *dbus.Signal, 50)

var properties *prop.Properties = nil
var	channels = make(ChannelMap)
var	mutex = sync.Mutex{}

type WatcherObject struct {}


func (wo WatcherObject) RegisterStatusNotifierItem(serviceName string) *dbus.Error {
	fmt.Println("Regster", serviceName)
	mutex.Lock()
	defer mutex.Unlock()

	if _, weHaveIt := channels[serviceName]; weHaveIt {
		return &dbus.Error{Name: "Already registered"}
	} else {
		channels[serviceName] = make(PropChangeChannel)
		go StatusNotifierItem(serviceName, channels[serviceName])
		conn.Emit(WATCHER_PATH, WATCHER_INTERFACE + ".StatusNotifierItemRegistered", serviceName)
		if props, err := getItemProps(serviceName); err == nil && len(props) > 0 {
			channels[serviceName] <- props
			updateWatcherProperties()
			return nil
		} else {
			fmt.Println("Error retrieving props:", err)
			return dbus.MakeFailedError(err)
		}

	}

}

func (wo* WatcherObject) UnregisterStatusNotifierItem(serviceName string) *dbus.Error {
	mutex.Lock()
	defer mutex.Unlock()

	if channel, ok := channels[serviceName]; ok {
		fmt.Println("UnregisterStatusNotifierItem: ", serviceName)
		close(channel)
		delete(channels, serviceName)
		conn.Emit(WATCHER_PATH, WATCHER_INTERFACE + ".StatusNotifierItemUnregistered", serviceName)
		updateWatcherProperties()
	}

	return nil
}


var itemFields = []string{
			"Id",
			"Category",
			"Status",
			"Title",
			"ItemIsMenu",
			"IconName",
			"AttentionIconName",
			"OverlayIconName",
			"AttentionMovieName",
			"IconPixmap",
			"AttentionIconPixmap",
			"OverlayIconPixmap",
//			"ToolTip",
}

func getItemProps(serviceName string) (Props, error) {
	obj := conn.Object(serviceName, dbus.ObjectPath(ITEM_PATH))
	props := make(Props)
	for _,itemFieldName := range itemFields {
		call := obj.Call("org.freedesktop.DBus.Properties.Get", dbus.Flags(0), ITEM_INTERFACE, itemFieldName)
		if call.Err == nil {
			props[itemFieldName] = call.Body[0].(dbus.Variant)
		} else {
			fmt.Println("Error getting", itemFieldName, call.Err)
		}
	}

	return props,nil
}

var watcherObject WatcherObject

// Caller must take mutex
func updateWatcherProperties() {
	items := make(stringlist.StringList, len(channels))
	pos := 0
	for serviceName := range channels {
		items[pos] = serviceName
		pos++
	}

	properties.Set(WATCHER_INTERFACE, "RegisteredStatusItems", dbus.MakeVariant(items))
}

func handleSignalFromItem(signal *dbus.Signal) {
	shortName := signal.Name[len("org.kde.StatusNotifierItem."):]
	fmt.Println("Signal from :", signal.Sender)
	switch (shortName) {
	case "NewTitle":
		fmt.Println("--> get title")
	case "NewIcon":
		fmt.Println("--> get icon")
	case "NewAttentionIcon":
		fmt.Println("--> get attentionIcon")
	case "NewOverlayIcon":
		fmt.Println("--> get overlayIcon")
	case "NewStatus":
		fmt.Println("--> get status")
	case "NewToolTip":
		fmt.Println("--> get tooltip")
	}
}

func runWatcher() {
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

	// Map dbus objects
	props := map[string]map[string]*prop.Prop{
		WATCHER_INTERFACE : {
				"IsStatusNotifierHostRegistered" : { true, false, prop.EmitTrue, nil},
				"ProtocolVersion" : {0, false, prop.EmitTrue, nil},
				"RegisteredStatusItems": { []string{}, false, prop.EmitTrue, nil},
		},
	}
	properties = prop.New(conn, WATCHER_PATH, props)
	conn.Export(watcherObject, WATCHER_PATH, WATCHER_INTERFACE)
	conn.Export(introspect.Introspectable(INTROSPECT_XML), WATCHER_PATH, INTROSPECT_INTERFACE)

	// Observe the world
	conn.Signal(dbusSignals)
	itemSignalsRule := "type='signal', interface='org.kde.StatusNotifierItem'"
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, itemSignalsRule)
	nameOwnerChangedRule := "type='signal', interface='org.freedesktop.DBus'"
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, nameOwnerChangedRule)

	fmt.Println("Waiting for signals")
	for signal := range dbusSignals {
		/*fmt.Print("signal: " + signal.Name + "(")
		sep := ""
		for _,inf := range signal.Body {
			fmt.Print(sep, "'", inf, "' ")
			sep = ", "
		}
		fmt.Print(")\n  sender: ", signal.Sender, "\n  path: ", signal.Path, "\n\n")
		*/

		if strings.HasPrefix(signal.Name, "org.kde.StatusNotifierItem.") {
			handleSignalFromItem(signal)

		} else if signal.Name == "org.freedesktop.DBus.NameOwnerChanged" && len(signal.Body) == 3 {
			arg0 := signal.Body[0].(string)
			arg1 := signal.Body[1].(string)
			arg2 := signal.Body[2].(string)

			if len(arg1) > 0 && len(arg2) == 0 {
				watcherObject.UnregisterStatusNotifierItem(arg0)
			}
		}
	}
}


