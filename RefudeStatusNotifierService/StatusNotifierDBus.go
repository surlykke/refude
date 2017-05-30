package main

import (
	"github.com/godbus/dbus"
	"errors"
	"fmt"
	"github.com/godbus/dbus/prop"
	"sync"
	"github.com/godbus/dbus/introspect"
	"github.com/surlykke/RefudeServices/lib/stringlist"
)

// Takes care of the dbus-side of things

const WATCHER_SERVICE = "org.kde.StatusNotifierWatcher"
const WATCHER_PATH = "/StatusNotifierWatcher"
const WATCHER_INTERFACE = WATCHER_SERVICE

const HOST_SERVICE = "org.kde.StatusNotifierHost"

const ITEM_PATH = "/StatusNotifierItem"
const ITEM_INTERFACE = "org.kde.StatusNotifierItem"

const ITEM_REGISTERED = WATCHER_INTERFACE + ".StatusNotifierItemRegistered"
const ITEM_UNREGISTERED = WATCHER_INTERFACE + ".StatusNotifierItemUnRegistered"
const HOST_REGISTERED = WATCHER_INTERFACE + ".StatusNotifierHostRegistered"

const INTROSPECT_INTERFACE = "org.freedesktop.DBus.Introspectable"
const PROPERTIES_INTERFACE = "org.freedesktop.DBus.Properties"

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
var registerChannel= make(chan string, 20)
var unregisterChannel = make(chan string, 20)
var dbusSignals = make(chan *dbus.Signal, 50)


type WatcherObject struct {
	properties *prop.Properties
	channels ChannelMap
	mutex sync.Mutex
}


func (wo WatcherObject) RegisterStatusNotifierItem(serviceName string) *dbus.Error {
	fmt.Println("Regster", serviceName)
	wo.mutex.Lock()
	defer wo.mutex.Unlock()

	fmt.Println("RegisterStatusNotifierItem: ", serviceName)
	if _, weHaveIt := wo.channels[serviceName]; weHaveIt {
		return &dbus.Error{Name: "Already registered"}
	} else {
		wo.channels[serviceName] = make(PropChangeChannel)
		listenForSignals(serviceName)
		go StatusNotifierItem(serviceName, wo.channels[serviceName])
		conn.Emit(WATCHER_PATH, WATCHER_INTERFACE + ".StatusNotifierItemRegistered", serviceName)
		if props, err := getItemProps(serviceName); err == nil && len(props) > 0 {
			wo.channels[serviceName] <- props
			registerChannel <- serviceName
			wo.updateProperties()
			return nil
		} else {
			fmt.Println("Error retrieving props:", err)
			return dbus.MakeFailedError(err)
		}

	}
}

func (wo WatcherObject) UnregisterStatusNotifierItem(serviceName string) *dbus.Error {
	wo.mutex.Lock()
	defer wo.mutex.Unlock()

	fmt.Println("RegisterStatusNotifierItem: ", serviceName)
	if channel, ok := wo.channels[serviceName]; ok {
		close(channel)
		delete(wo.channels, serviceName)
		conn.Emit(WATCHER_PATH, WATCHER_INTERFACE + ".StatusNotifierItemUnregistered", serviceName)
		unregisterChannel <- serviceName
		wo.updateProperties()
	}

	return nil
}

func (wo WatcherObject) RegisterStatusNotifierHost(service string) {
}

var watcherObject WatcherObject


func SetupWatcherObject() {
	watcherObject = WatcherObject{nil, make(ChannelMap), sync.Mutex{}}
	props := map[string]map[string]*prop.Prop{
		WATCHER_INTERFACE : {
				"IsStatusNotifierHostRegistered" : { true, false, prop.EmitTrue, nil},
				"ProtocolVersion" : {0, false, prop.EmitTrue, nil},
				"RegisteredStatusItems": { []string{}, false, prop.EmitTrue, nil},
		},
	}
	watcherObject.properties = prop.New(conn, WATCHER_PATH, props)
	conn.Export(watcherObject, WATCHER_PATH, WATCHER_INTERFACE)
	conn.Export(introspect.Introspectable(INTROSPECT_XML), WATCHER_PATH, INTROSPECT_INTERFACE)

}



func init() {
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

	SetupWatcherObject()

	conn.Signal(dbusSignals)
}

func listenForSignals(serviceName string) {
	rule := "type='signal'," +
			"interface='org.freedesktop.DBus.Properties'," +
			"member='PropertiesChanged'," +
			"sender='" + serviceName + "'"

	conn.BusObject().Call( "org.freedesktop.DBus.AddMatch", 0, rule)
}


// Caller must take mutex
func (wo WatcherObject) updateProperties() {
	items := make(stringlist.StringList, len(wo.channels))
	pos := 0
	for serviceName := range wo.channels {
		items[pos] = serviceName
		pos++
	}

	wo.properties.Set(WATCHER_INTERFACE, "RegisteredStatusItems", dbus.MakeVariant(items))
}

func consumeSignals() {
	fmt.Println("Waiting for signals")
	for signal := range dbusSignals {
		watcherObject.dispatch(signal)
	}
}

func (wo WatcherObject) dispatch(signal *dbus.Signal) {
	wo.mutex.Lock()
	defer wo.mutex.Unlock()

	if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
		if channel, ok := wo.channels[signal.Sender]; ok {
			channel <- signal.Body[0].(map[string]dbus.Variant)
		}
	}
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
