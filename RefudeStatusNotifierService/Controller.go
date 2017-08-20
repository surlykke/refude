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
var	channels = make(map[SenderAndPath]chan *dbus.Signal)
var	mutex = sync.Mutex{}

var watcherProperties *prop.Properties

func serviceKey(sender dbus.Sender, objectPath dbus.ObjectPath) string {
	return string(sender) + string(objectPath)
}

func restPath(sender dbus.Sender, objectPath dbus.ObjectPath) string {
	return "/items/" + strings.Replace(serviceKey(sender, objectPath)[1:], "/", "-", -1)
}

func makeSenderAndPath(serviceName string, sender dbus.Sender) SenderAndPath {
	var sp = SenderAndPath{sender: string(sender)}
	if regexp.MustCompile("^(/\\w+)+$").MatchString(serviceName) {
		sp.objPath = dbus.ObjectPath(serviceName)
	} else {
		sp.objPath = dbus.ObjectPath(ITEM_PATH)
	}
	return sp
}

/**
 * serviceId Can be a name of service or a path of object
 */
func addItem(serviceName string, sender dbus.Sender) *dbus.Error {
	var sp = makeSenderAndPath(serviceName, sender)

	mutex.Lock()
	defer mutex.Unlock()
	if _,exists := channels[sp]; !exists {
		var item = MakeItem(sp)
		channels[sp] = make(chan *dbus.Signal)
		var path = "/items/" + sp.sender[1:] + strings.Replace(string(sp.objPath), "/", "-", -1)
		go item.Run(path, channels[sp])
		watcherProperties.Set(WATCHER_INTERFACE, "RegisteredStatusItems", dbus.MakeVariant(getItems()))
		conn.Emit(WATCHER_PATH, WATCHER_INTERFACE + ".StatusNotifierItemRegistered", string(sender))
		return nil
	} else {
		return dbus.MakeFailedError(errors.New("Already registered"))
	}
}

func removeItem(serviceName string, sender dbus.Sender) {
	var sp = makeSenderAndPath(serviceName, sender)

	mutex.Lock()
	defer mutex.Unlock()
	if channel, ok := channels[sp]; ok {
		close(channel)
		delete(channels, sp)
		watcherProperties.Set(WATCHER_INTERFACE, "RegisteredStatusItems", dbus.MakeVariant(getItems()))
	}
}

func removeSender(sender dbus.Sender) {
	fmt.Println("Remove sender: ", sender)
	mutex.Lock()
	defer mutex.Unlock()

	var somethingRemoved = false
	for sp, channel := range channels {
		fmt.Println("Looking at:", sp)
		if sp.sender == string(sender) {
			fmt.Println("removing")
			close(channel)
			delete(channels, sp)
			somethingRemoved = true
		}
	}

	if somethingRemoved {
		watcherProperties.Set(WATCHER_INTERFACE, "RegisteredStatusItems", dbus.MakeVariant(getItems()))
	}
}


// Caller holds mutex
func getItems() []string {
	res := make([]string, 0)
	for sp,_ := range channels {
		res = append(res, sp.sender + ":" + string(sp.objPath)) // FIXME
	}
	return res
}

func dispatchSignal(signal *dbus.Signal) {
	var sp = SenderAndPath{signal.Sender, signal.Path}
	mutex.Lock()
	defer mutex.Unlock()
	if channel, ok := channels[sp]; ok {
		channel <- signal
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
			fmt.Println("NameOwnerChanged: ", signal.Body)
			arg0 := signal.Body[0].(string)
			arg1 := signal.Body[1].(string)
			arg2 := signal.Body[2].(string)
			if len(arg1) > 0 && len(arg2) == 0 { // Someone had the name and now no-one does
												 // We take that to mean that the app has exited
				removeSender(dbus.Sender(arg0))
			}
		}
	}
}


