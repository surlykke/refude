// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package power

import (
	"github.com/godbus/dbus/v5"
	dbuscall "github.com/surlykke/RefudeServices/lib/dbusutils"
)

const upowerService = "org.freedesktop.UPower"
const upowerPath = "/org/freedesktop/UPower"
const upowerInterface = "org.freedesktop.UPower"
const devicePrefix = "/org/freedesktop/UPower/devices"
const displayDeviceDbusPath = dbus.ObjectPath(devicePrefix + "/DisplayDevice")
const upowerDeviceInterface = "org.freedesktop.UPower.Device"
const displayDevicePath = "/device/DisplayDevice"

func subscribe() chan *dbus.Signal {
	var signals = make(chan *dbus.Signal, 100)

	dbusConn.Signal(signals)
	dbusConn.BusObject().Call(
		"org.freedesktop.DBus.AddMatch",
		0,
		"type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged', sender='org.freedesktop.UPower'")

	dbusConn.BusObject().Call(
		"org.freedesktop.DBus.AddMatch",
		0,
		"type='signal',interface='org.freedesktop.UPower',member='DeviceAdded', sender='org.freedesktop.UPower'")

	dbusConn.BusObject().Call(
		"org.freedesktop.DBus.AddMatch",
		0,
		"type='signal',interface='org.freedesktop.UPower',member='DeviceRemoved', sender='org.freedesktop.UPower'")

	return signals
}

func retrieveDevicePaths() []dbus.ObjectPath {
	enumCall := dbusConn.Object(upowerService, upowerPath).Call(upowerInterface+".EnumerateDevices", dbus.Flags(0))
	return append(enumCall.Body[0].([]dbus.ObjectPath), displayDeviceDbusPath)
}

func retrieveDevice(path dbus.ObjectPath) *Device {
	var device = Device{
		DbusPath:      path,
		DisplayDevice: path == displayDeviceDbusPath,
	}

	var props = dbuscall.GetAllProps(dbusConn, upowerService, path, upowerDeviceInterface)

	for key, variant := range props {
		switch key {
		case "NativePath":
			device.NativePath = variant.Value().(string)
		case "Vendor":
			device.Vendor = variant.Value().(string)
		case "Model":
			device.Model = variant.Value().(string)
		case "Serial":
			device.Serial = variant.Value().(string)
		case "UpdateTime":
			device.UpdateTime = variant.Value().(uint64)
		case "Type":
			device.Type = deviceType(variant.Value().(uint32))
		case "PowerSupply":
			device.PowerSupply = variant.Value().(bool)
		case "HasHistory":
			device.HasHistory = variant.Value().(bool)
		case "HasStatistics":
			device.HasStatistics = variant.Value().(bool)
		case "Online":
			device.Online = variant.Value().(bool)
		case "Energy":
			device.Energy = variant.Value().(float64)
		case "EnergyEmpty":
			device.EnergyEmpty = variant.Value().(float64)
		case "EnergyFull":
			device.EnergyFull = variant.Value().(float64)
		case "EnergyFullDesign":
			device.EnergyFullDesign = variant.Value().(float64)
		case "EnergyRate":
			device.EnergyRate = variant.Value().(float64)
		case "Voltage":
			device.Voltage = variant.Value().(float64)
		case "TimeToEmpty":
			device.TimeToEmpty = variant.Value().(int64)
		case "TimeToFull":
			device.TimeToFull = variant.Value().(int64)
		case "Percentage":
			device.Percentage = int8(variant.Value().(float64))
		case "IsPresent":
			device.IsPresent = variant.Value().(bool)
		case "State":
			device.State = deviceState(variant.Value().(uint32))
		case "IconName":
			device.IconName = variant.Value().(string)
		case "IsRechargeable":
			device.IsRechargeable = variant.Value().(bool)
		case "Capacity":
			device.Capacity = variant.Value().(float64)
		case "Technology":
			device.Technology = deviceTecnology(variant.Value().(uint32))
		}
	}
	device.title = deviceTitle(device.Type, device.Model)
	return &device
}

var dbusConn = func() *dbus.Conn {
	if conn, err := dbus.SystemBus(); err != nil {
		panic(err)
	} else {
		return conn
	}
}()

func updateDevice(d *Device, m map[string]dbus.Variant) {

}
