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
	"github.com/godbus/dbus/introspect"
	"strings"
	"github.com/godbus/dbus/prop"
	"regexp"
	"github.com/surlykke/RefudeServices/lib/service"
)

const WATCHER_SERVICE = "org.kde.StatusNotifierWatcher"
const WATCHER_PATH = "/StatusNotifierWatcher"
const WATCHER_INTERFACE = WATCHER_SERVICE
const HOST_SERVICE = "org.kde.StatusNotifierHost"
const ITEM_PATH = "/StatusNotifierItem"
const ITEM_INTERFACE = "org.kde.StatusNotifierItem"
const INTROSPECT_INTERFACE = "org.freedesktop.DBus.Introspectable"
const PROPERTIES_INTERFACE = "org.freedesktop.DBus.Properties"
const MENU_INTERFACE = "com.canonical.dbusmenu"

var	conn *dbus.Conn
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
	var event = Event{eventType: ItemCreated}
	event.sender, event.path = senderAndPath(serviceName, sender)
	events <- event
	return nil
}

func removeItem(serviceName string, sender dbus.Sender) {
	var event = Event{eventType: ItemRemoved}
	event.sender, event.path = senderAndPath(serviceName, sender)
	events <- event
}




func monitorSignals() {
	var dbusSignals = make(chan *dbus.Signal, 50)
	conn.Signal(dbusSignals)

	itemSignalsRule := "type='signal', interface='org.kde.StatusNotifierItem'"
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, itemSignalsRule)

	nameOwnerChangedRule := "type='signal', interface='org.freedesktop.DBus'"
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, nameOwnerChangedRule)

	fmt.Println("Waiting for signals")
	for signal := range dbusSignals {
		if strings.HasPrefix(signal.Name, "org.kde.StatusNotifierItem.New") {
			events <- Event{eventType: ItemUpdated, sender: signal.Sender, path: signal.Path}
		} else if signal.Name == "org.freedesktop.DBus.NameOwnerChanged" && len(signal.Body) == 3 {
			fmt.Printf("NameOwnerChanged: '%s', '%s', '%s'\n", signal.Body[0], signal.Body[1], signal.Body[2])
			sender := signal.Body[0].(string)
			oldOwner := signal.Body[1].(string)
			newOwner := signal.Body[2].(string)
			if len(oldOwner) > 0 && len(newOwner) == 0 { // Someone had the name and now no-one does
												 // We take that to mean that the app has exited
				events <- Event{eventType:SenderTerminated, sender: sender}
			}
		}
	}
}

//


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
	conn.ExportMethodTable(
		map[string]interface{}{
			"RegisterStatusNotifierItem": addItem,
			"UnregisterStatusNotifierItem": removeItem,
		},
		WATCHER_PATH,
		WATCHER_INTERFACE,
	)

	// Add Introspectable interface
	conn.Export(introspect.Introspectable(INTROSPECT_XML), WATCHER_PATH, INTROSPECT_INTERFACE)

	// Add properties interface
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
}


var items = []*Item{}

type EventType int

const (
	ItemCreated EventType = iota
	ItemRemoved
	SenderTerminated
	ItemUpdated
	MenuUpdated
)

type Event struct {
	eventType EventType
	sender string
	path   dbus.ObjectPath // Menupath if eventType is MenuUpdated, ItemPath otherwise
}

var events = make(chan Event)

func findByItemPath(sender string, itemPath dbus.ObjectPath) int {
	for i,item := range items {
		if sender == item.sender && itemPath == item.itemPath {
			return i
		}
	}
	return -1
}

func findByMenuPath(sender string, menuPath dbus.ObjectPath) int {
	for i,item := range items {
		if sender == item.sender && menuPath == item.menuPath {
			return i
		}
	}
	return -1
}

func updateWatcherProperties() {
	ids := make([]string,0, len(items))
	for _,item := range items {
		ids = append(ids, item.sender + ":" + string(item.itemPath))
	}
	watcherProperties.Set(WATCHER_INTERFACE, "RegisteredStatusItems", dbus.MakeVariant(ids))
}

func Controller() {
	getOnTheBus()
	go monitorSignals()

	for event := range events {
		switch event.eventType {
		case ItemCreated:
			if findByItemPath(event.sender, event.path) == -1 {
				item := &Item{sender: event.sender, itemPath: event.path}
				item.fetchProps()
				if item.menuPath != "" {
					item.fetchMenu()
				}
				items = append(items, item)
				service.Map(item.restPath(), item.copy())
				updateWatcherProperties()
			}
		case ItemRemoved:
			if index := findByItemPath(event.sender, event.path); index > -1 {
				service.Unmap(items[index].restPath())
				items = append(items[0:index], items[index + 1:len(items)]...)
				updateWatcherProperties()
			}
		case SenderTerminated:
			tmp := []*Item{}
			for _,item := range items {
				if event.sender == item.sender {
					service.Unmap(item.restPath())
				} else {
					tmp = append(tmp, item)
				}
			}
			if len(tmp) != len(items) {
				items = tmp
				updateWatcherProperties()
			}
		case ItemUpdated:
			if index := findByItemPath(event.sender, event.path); index > -1 {
				items[index].fetchProps()
				service.Map(items[index].restPath(), items[index].copy())
			}
		case MenuUpdated:
			if index := findByMenuPath(event.sender, event.path); index > -1 {
				items[index].fetchMenu()
				service.Map(items[index].restPath(), items[index].copy())
			}
		}
	}
}
