package main

import (
	"github.com/godbus/dbus"
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
		service.Remap("/UPower", uPower) // Do
	}
}

// Keeps an eye on a device. PowerManager#Run redirects PropertiesChanged to this through the changes channel
func watchDevice(conn *dbus.Conn, dbusPath dbus.ObjectPath, changes chan map[string]dbus.Variant) {
	path := resourcePath(dbusPath)
	device := Device{}
	device.ReadDBusProps(getProps(conn, dbusPath, UPowerDeviceInterface))
	service.Map(path, device) // Important to call-by-value (copy) here. Service owns what it gets
	for {
		device.ReadDBusProps(<-changes)
		service.Remap(path, device) // do.
	}
}

func resourcePath(devicePath dbus.ObjectPath) string {
	return "/device/" + string(devicePath)[strings.LastIndex(string(devicePath), "/") + 1:]
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
	for _,devicePath := range devicePaths {
		pm.changeChans[devicePath] = make(chan map[string]dbus.Variant)
		go watchDevice(pm.conn, devicePath, pm.changeChans[devicePath])
	}

	resourcePaths := make(common.StringSet)
    for _,devicePath := range devicePaths {
	    resourcePaths[resourcePath(devicePath)[1:]] = true
    }
	service.Map("/devices", &resourcePaths)

	for signal := range signals {
		if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			if changeChan, ok := pm.changeChans[signal.Path]; ok {
				changeChan <- signal.Body[1].(map[string]dbus.Variant)
			}
		}
	}
}
