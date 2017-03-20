package main

import (
	"github.com/godbus/dbus"
	"fmt"
	"github.com/surlykke/RefudeServices/service"
	"github.com/surlykke/RefudeServices/common"
	"strings"
)

const UPowService = "org.freedesktop.UPower"
const UPowPath = "/org/freedesktop/UPower"
const UPowerInterface = "org.freedesktop.UPower"
const DisplayDevicePath = "/org/freedesktop/UPower/devices/DisplayDevice"
const IntrospectInterface = "org.freedesktop.DBus.Introspectable"
const DBusPropertiesInterface = "org.freedesktop.DBus.Properties"

const UPowerDeviceInterface = "org.freedesktop.UPower.Device"


type PowerManager struct {
	watchers map[dbus.ObjectPath]chan bool
	conn *dbus.Conn
}


// Keeps an eye on UPower. Send true through nudges to make it refresh, false to return
func watchUPower(conn *dbus.Conn, dbusPath dbus.ObjectPath, nudges chan bool) {
	service.Map("/UPower", &UPower{})
	for {
		if ! <-nudges {
			return
		}

		uPower := UPower{}
		uPower.ReadDBusProps(getProps(conn, dbusPath, UPowerInterface))
		service.Remap("/UPower", &uPower)
	}
}


// Keeps an eye on a device. Send true through nudges to make it refresh, false to return
func watchDevice(conn *dbus.Conn, dbusPath dbus.ObjectPath, resourcePath string, nudges chan bool) {
	service.Map(resourcePath, &Device{})
	for {
		if ! <-nudges {
			return
		}

		device := Device{}
		device.ReadDBusProps(getProps(conn, dbusPath, UPowerDeviceInterface))
		service.Remap(resourcePath, &device)
	}
}

func getProps(conn *dbus.Conn, path dbus.ObjectPath, dbusInterface string) map[string]dbus.Variant {
	call :=  conn.Object(UPowService, path).Call("org.freedesktop.DBus.Properties.GetAll", dbus.Flags(0), dbusInterface)
	return call.Body[0].(map[string]dbus.Variant)
}

func (pm *PowerManager) listenForPropChanges(path dbus.ObjectPath) {
	matchRule := fmt.Sprintf("type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged',path='%s'", path)
	fmt.Println("MatchRule: ", matchRule)
	call := pm.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, matchRule)
	fmt.Println(".. -> ", call)
}

func (pm *PowerManager) Run() {
	pm.watchers = make(map[dbus.ObjectPath]chan bool)

	if conn, err := dbus.SystemBus(); err != nil {
		panic(err)
	} else {
		pm.conn = conn
	}

	signals := make(chan *dbus.Signal, 100)
	pm.conn.Signal(signals)


	pm.listenForPropChanges(UPowPath)
	pm.watchers[UPowPath] = make(chan bool)
	go watchUPower(pm.conn, UPowPath, pm.watchers[UPowPath])


	uPowerObject := pm.conn.Object(UPowService, UPowPath)
	devicePaths := make(common.StringSet)

	enumCall := uPowerObject.Call(UPowerInterface + ".EnumerateDevices", dbus.Flags(0))

	for _,devicePath := range append(enumCall.Body[0].([]dbus.ObjectPath), DisplayDevicePath) {
		pm.listenForPropChanges(devicePath)
		pm.watchers[devicePath] = make(chan bool)
		resourcePath := "/device/" + string(devicePath)[strings.LastIndex(string(devicePath), "/") + 1:]
		go watchDevice(pm.conn, devicePath, resourcePath, pm.watchers[devicePath])
		devicePaths[resourcePath[1:]] = true
	}

	service.Map("/devices", &devicePaths)

	for signal := range signals {
		fmt.Println(signal.Path)
		if watcher,ok := pm.watchers[signal.Path]; ok {
			fmt.Println("enqueing")
			watcher <- true
			fmt.Println("enqueued")
		}
	}
}
