// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/godbus/dbus"
	"errors"
	"fmt"
	"sync"
	"github.com/godbus/dbus/introspect"
	"strings"
	"github.com/godbus/dbus/prop"
	"regexp"
)

const WATCHER_SERVICE = "org.kde.StatusNotifierWatcher"
const WATCHER_PATH = "/StatusNotifierWatcher"
const WATCHER_INTERFACE = WATCHER_SERVICE
const HOST_SERVICE = "org.kde.StatusNotifierHost"
const ITEM_PATH = "/StatusNotifierItem"
const ITEM_INTERFACE = "org.kde.StatusNotifierItem"
const INTROSPECT_INTERFACE = "org.freedesktop.DBus.Introspectable"
const PROPERTIES_INTERFACE = "org.freedesktop.DBus.Properties"

var	conn *dbus.Conn
var dbusSignals = make(chan *dbus.Signal, 50)
var	channels = make(map[string]chan string)
var	mutex = sync.Mutex{}

var watcherProperties *prop.Properties


func GetNameOwner(serviceName string) (string, error) {
	call := conn.BusObject().Call("GetNameOwner", dbus.Flags(0), serviceName)
	return call.Body[0].(string), call.Err
}

func addItem(serviceName string, sender dbus.Sender) *dbus.Error {
	// Problem: We cant handle more than one item pr sender. Not sure
	// how to do that with the spec as it is
	fmt.Println("serviceName: ", serviceName, ", sender: ", sender)
	serviceOwner := string(sender)
	var objectPath dbus.ObjectPath = ""
	if regexp.MustCompile("^(/\\w+)+$").MatchString(serviceName) {
		objectPath = dbus.ObjectPath(serviceName)
	} else {
		objectPath = dbus.ObjectPath("/StatusNotifierItem")
	}

	mutex.Lock()
	defer mutex.Unlock()

	if _,exists := channels[serviceOwner]; !exists {
		channels[serviceOwner] = make(chan string)
		go StatusNotifierItem(serviceOwner, objectPath, channels[serviceOwner])
		watcherProperties.Set(WATCHER_INTERFACE, "RegisteredStatusItems", dbus.MakeVariant(getItems()))
		conn.Emit(WATCHER_PATH, WATCHER_INTERFACE + ".StatusNotifierItemRegistered", serviceOwner)
		return nil
	} else {
		return dbus.MakeFailedError(errors.New("Already registered"))
	}
}

func removeItem(serviceName string) *dbus.Error {
	mutex.Lock()
	defer mutex.Unlock()

	if channel, ok := channels[serviceName]; ok {
		close(channel)
		delete(channels, serviceName)
		watcherProperties.Set(WATCHER_INTERFACE, "RegisteredStatusItems", dbus.MakeVariant(getItems()))
		return nil
	} else {
		return dbus.MakeFailedError(errors.New("Unknown item"))
	}
}

// Caller holds mutex
func getItems() []string {
	res := make([]string, 0)
	for serviceName,_ := range channels {
		res = append(res, serviceName)
	}
	return res
}

func dispatchSignal(signal *dbus.Signal) {
	shortName := signal.Name[len("org.kde.StatusNotifierItem."):]
	mutex.Lock()
	defer mutex.Unlock()
	if channel, ok := channels[signal.Sender]; ok {
		channel <- shortName
	}

}

func run() {
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
	conn.ExportMethodTable(
		map[string]interface{}{ "RegisterStatusNotifierItem": addItem, "UnregisterStatusNotifierItem": removeItem},
		WATCHER_PATH,
		WATCHER_INTERFACE,
	)
	conn.Export(introspect.Introspectable(INTROSPECT_XML), WATCHER_PATH, INTROSPECT_INTERFACE)
    watcherProperties = prop.New(
		conn,
		WATCHER_PATH,
		map[string]map[string]*prop.Prop{
			WATCHER_INTERFACE: {
				"IsStatusNotifierHostRegistered": {true, false, prop.EmitTrue, nil},
				"ProtocolVersion":                {0, false, prop.EmitTrue, nil},
				"RegisteredStatusItems":          {[]string{}, false, prop.EmitTrue, nil},
			},
		},
	)


	// Consume signals
	conn.Signal(dbusSignals)

	itemSignalsRule := "type='signal', interface='org.kde.StatusNotifierItem'"
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, itemSignalsRule)

	nameOwnerChangedRule := "type='signal', interface='org.freedesktop.DBus'"
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, nameOwnerChangedRule)

	fmt.Println("Waiting for signals")
	for signal := range dbusSignals {
		if strings.HasPrefix(signal.Name, "org.kde.StatusNotifierItem.") {
			dispatchSignal(signal)

		} else if signal.Name == "org.freedesktop.DBus.NameOwnerChanged" && len(signal.Body) == 3 {
			arg0 := signal.Body[0].(string)
			arg1 := signal.Body[1].(string)
			arg2 := signal.Body[2].(string)
			if len(arg1) > 0 && len(arg2) == 0 { // Someone had the name and now no-one does
												 // We take that to mean that the app has exited
				removeItem(arg0)
			}
		}
	}
}


