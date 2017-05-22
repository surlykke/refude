/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package main

import (
	"github.com/godbus/dbus"
	"github.com/surlykke/RefudeServices/service"
	"github.com/surlykke/RefudeServices/common"
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

type PowerManager struct {
	changeChans map[dbus.ObjectPath]chan map[string]dbus.Variant
	conn *dbus.Conn
}

// Keeps an eye on UPower. PowerManager#Run redirects PropertiesChanged to this through the changes channel
func watchUPower(conn *dbus.Conn, changes chan map[string]dbus.Variant) {
	uPower := UPower{}
	uPower.ReadDBusProps(getProps(conn, UPowPath, UPowerInterface))
	service.Map("/UPower", uPower) // Important that we use call-by-value, ie. uPower is copied - see below
	for {
		uPower.ReadDBusProps(<- changes)
		service.Map("/UPower", uPower) // Do
	}
}

// Keeps an eye on a device. PowerManager#Run redirects PropertiesChanged to this through the changes channel
func watchDevice(conn *dbus.Conn, dbusPath dbus.ObjectPath, changes chan map[string]dbus.Variant) {
	path := "/devices/" + resourcePath(dbusPath)
	device := Device{}
	device.ReadDBusProps(getProps(conn, dbusPath, UPowerDeviceInterface))
	service.Map(path, device) // Important to call-by-value (copy) here. Service owns what it gets
	for {
		device.ReadDBusProps(<-changes)
		service.Map(path, device) // do.
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

	if conn, err := dbus.SystemBus(); err != nil {
		panic(err)
	} else {
		pm.conn = conn
	}

	signals := make(chan *dbus.Signal, 100)
	pm.conn.Signal(signals)
	pm.conn.BusObject().Call(
		"org.freedesktop.DBus.AddMatch",
		0,
		"type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged', sender='org.freedesktop.UPower'")


	pm.changeChans[UPowPath] = make(chan map[string]dbus.Variant)
	go watchUPower(pm.conn, pm.changeChans[UPowPath])

	enumCall := pm.conn.Object(UPowService, UPowPath).Call(UPowerInterface + ".EnumerateDevices", dbus.Flags(0))
	devicePaths := append(enumCall.Body[0].([]dbus.ObjectPath), DisplayDevicePath)
	resourcePaths := make(common.StringList, 0, len(devicePaths))
	for _,devicePath := range devicePaths {
		resourcePaths = append(resourcePaths, resourcePath(devicePath))
		pm.changeChans[devicePath] = make(chan map[string]dbus.Variant)
		go watchDevice(pm.conn, devicePath, pm.changeChans[devicePath])
	}

	service.Map("/devices/", resourcePaths)

	actions := []*PowerAction{
		NewPowerAction("PowerOff", "Shutdown", "Power off the machine", "system-shutdown", pm.conn),
		NewPowerAction("Reboot", "Reboot", "Reboot the machine", "system-reboot", pm.conn),
		NewPowerAction("Suspend", "Suspend", "Suspend the machine", "system-suspend", pm.conn),
		NewPowerAction("Hibernate", "Hibernate", "Put the machine into hibernation", "system-suspend-hibernate", pm.conn),
		NewPowerAction("HybridSleep", "HybridSleep", "Put the machine into hybrid sleep", "system-suspend-hibernate", pm.conn),
	}
	actionIds := make(common.StringList, len(actions))
	for i,action := range(actions) {
		actionIds[i] = action.Id
		fmt.Println("Mapping ", "/actions/" + action.Id)
		service.Map("/actions/" + action.Id, action)
	}
	service.Map("/actions/", &actionIds)

	service.Map("/", common.StringList{"ping", "notify", "UPower", "actions/", "devices/"})
	for signal := range signals {
		if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			if changeChan, ok := pm.changeChans[signal.Path]; ok {
				changeChan <- signal.Body[1].(map[string]dbus.Variant)
			}
		}
	}
}


