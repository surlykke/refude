// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/godbus/dbus"
	"github.com/surlykke/RefudeServices/lib/service"
	"strings"
	"fmt"
)

const UPowService = "org.freedesktop.UPower"
const UPowPath = "/org/freedesktop/UPower"
const UPowerInterface = "org.freedesktop.UPower"
const DisplayDevicePath = "/org/freedesktop/UPower/devices/DisplayDevice"
const IntrospectInterface = "org.freedesktop.DBus.Introspectable"
const DBusPropertiesInterface = "org.freedesktop.DBus.Properties"
const UPowerDeviceInterface = "org.freedesktop.UPower.Device"
const login1Service = "org.freedesktop.login1"
const login1Path = "/org/freedesktop/login1"
const managerInterface = "org.freedesktop.login1.Manager"

var dbusConn = 	func() *dbus.Conn {
					if conn, err := dbus.SystemBus(); err != nil {
						panic(err)
					} else {
						return conn
					}
				}()


type PowerManager struct {
	changeChans map[dbus.ObjectPath]chan map[string]dbus.Variant
}

// Keeps an eye on UPower. PowerManager#Run redirects PropertiesChanged to this through the changes channel
func watchUPower(changes chan map[string]dbus.Variant) {
	uPower := UPower{}
	uPower.ReadDBusProps(getProps(dbusConn, UPowPath, UPowerInterface))

    // Important that we use a copy - see below
	copy := uPower
	service.Map( "/UPower", &copy)

	for {
		uPower.ReadDBusProps(<- changes)
		copy := uPower
		service.Map( "/UPower", &copy)
	}
}

// Keeps an eye on a device.
// PowerManager#Run redirects PropertiesChanged to this through the changes channel
func watchDevice(dbusPath dbus.ObjectPath, changes chan map[string]dbus.Variant) {
	path := "/devices/" + resourcePath(dbusPath)
	device := Device{DisplayDevice: "/devices/DisplayDevice" == path}
	device.ReadDBusProps(getProps(dbusConn, dbusPath, UPowerDeviceInterface))
	// Important to use copy here. Service owns what it gets
	copy := device
	service.Map( path, &copy)

	for {
		device.ReadDBusProps(<-changes)
		copy := device
		service.Map(path, &copy)
	}
}

func resourcePath(devicePath dbus.ObjectPath) string {
	res := string(devicePath)[strings.LastIndex(string(devicePath), "/") + 1:]
	return res
}

func getProps(conn *dbus.Conn, path dbus.ObjectPath, dbusInterface string) map[string]dbus.Variant {
	call :=  conn.Object(UPowService, path).Call("org.freedesktop.DBus.Properties.GetAll", dbus.Flags(0), dbusInterface)
	return call.Body[0].(map[string]dbus.Variant)
}


func (pm *PowerManager) listenForPropChanges(path dbus.ObjectPath) {

}

func (pm *PowerManager) Run() {
	pm.changeChans = make(map[dbus.ObjectPath]chan map[string]dbus.Variant)



	signals := make(chan *dbus.Signal, 100)
	dbusConn.Signal(signals)
	dbusConn.BusObject().Call(
		"org.freedesktop.DBus.AddMatch",
		0,
		"type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged', sender='org.freedesktop.UPower'")


	pm.changeChans[UPowPath] = make(chan map[string]dbus.Variant)
	go watchUPower(pm.changeChans[UPowPath])

	enumCall := dbusConn.Object(UPowService, UPowPath).Call(UPowerInterface + ".EnumerateDevices", dbus.Flags(0))
	devicePaths := append(enumCall.Body[0].([]dbus.ObjectPath), DisplayDevicePath)
	for _,devicePath := range devicePaths {
		pm.changeChans[devicePath] = make(chan map[string]dbus.Variant)
		go watchDevice(devicePath, pm.changeChans[devicePath])
	}

	actions := []*PowerAction{
		NewPowerAction("PowerOff", "Shutdown", "Power off the machine", "system-shutdown"),
		NewPowerAction("Reboot", "Reboot", "Reboot the machine", "system-reboot"),
		NewPowerAction("Suspend", "Suspend", "Suspend the machine", "system-suspend"),
		NewPowerAction("Hibernate", "Hibernate", "Put the machine into hibernation", "system-suspend-hibernate"),
		NewPowerAction("HybridSleep", "HybridSleep", "Put the machine into hybrid sleep", "system-suspend-hibernate"),
	}

	for _,action := range(actions) {
		action.Self = "power-service:/actions/" + action.Id
		service.Map("/actions/" + action.Id, action)
	}

	for signal := range signals {
		if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			fmt.Println("Incoming: ", signal)
			if changeChan, ok := pm.changeChans[signal.Path]; ok {
				changeChan <- signal.Body[1].(map[string]dbus.Variant)
			}
		}
	}
}

var matchFunction service.MatchFunction = func(key string, value string, resource interface{}) bool {
	if key == "q" {
		if _, isDevice := resource.(*Device); isDevice {
			return true
		} else if action, isAction := resource.(*PowerAction); isAction {
			return strings.Contains(strings.ToUpper(action.Name), strings.ToUpper(value)) ||
				strings.Contains(strings.ToUpper(action.Comment), strings.ToUpper(value))
		}
	} else if key == "type" {
		if value == "ACTION" {
			_,ok := resource.(*PowerAction)
			return ok
		} else if value == "DEVICE" {
			_,ok := resource.(*Device)
			return ok
		}
	}

	return false
}

