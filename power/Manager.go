// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package power

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	dbuscall "github.com/surlykke/RefudeServices/lib/dbusutils"
	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/path"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/notifications"
)

const upowerService = "org.freedesktop.UPower"
const upowerPath = "/org/freedesktop/UPower"
const upowerInterface = "org.freedesktop.UPower"
const devicePrefix = "/org/freedesktop/UPower/devices/"
const displayDeviceDbusPath = dbus.ObjectPath(devicePrefix + "DisplayDevice")
const upowerDeviceInterface = "org.freedesktop.UPower.Device"

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

func retrieveDevice(dbusPath dbus.ObjectPath) *Device {
	var device = Device{
		ResourceData:  *resource.MakeBase(path.Of("/device/", dbusPath2id(dbusPath)), "", "", "", mediatype.Device),
		DbusPath:      dbusPath,
		DisplayDevice: dbusPath == displayDeviceDbusPath,
	}

	var props = dbuscall.GetAllProps(dbusConn, upowerService, dbusPath, upowerDeviceInterface)

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
		case "Luminosity":
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
		case "IsRechargeable":
			device.IsRechargeable = variant.Value().(bool)
		case "Capacity":
			device.Capacity = variant.Value().(float64)
		case "Technology":
			device.Technology = deviceTecnology(variant.Value().(uint32))
		case "WarningLevel":
			device.Warninglevel = deviceWarningLevel(variant.Value().(uint32))
		case "BatteryLevel":
			device.Batterylevel = deviceBatteryLevel(variant.Value().(uint32))
		case "IconName":
			device.Icon = icon.Name((variant.Value().(string)))
		}
	}
	device.Title = deviceTitle(device.Type, device.Model)
	device.Keywords = []string{"battery"}
	return &device
}

var previousPercentage = 101

// Sufficiently random
const notificationId uint32 = 1152165262

func showOnDesktop() {
	notifyOnLow()
	updateTrayIcon()
}

func updateTrayIcon() {
	// TODO
}

func notifyOnLow() {
	if displayDevice, ok := repo.Get[*Device](path.Of("/device/%s", dbusPath2id(displayDeviceDbusPath))); ok {
		var percentage = int(displayDevice.Percentage)
		if displayDevice.State == "Discharging" {
			if percentage <= 5 {
				notifications.Notify("refude", notificationId, "dialog-warning", "Battery critical", fmt.Sprintf("At %d%%", percentage), []string{}, map[string]dbus.Variant{"urgency": dbus.MakeVariant(uint8(2))}, -1)
			} else if percentage <= 10 && previousPercentage > 10 {
				notifications.Notify("refude", notificationId, "dialog-information", "Battery", fmt.Sprintf("At %d%%", percentage), []string{}, map[string]dbus.Variant{}, 10000)
			} else if percentage <= 15 && previousPercentage > 15 {
				notifications.Notify("refude", notificationId, "dialog-information", "Battery", fmt.Sprintf("At %d%%", percentage), []string{}, map[string]dbus.Variant{}, 5000)
			}
			previousPercentage = percentage
		} else {
			if percentage <= 15 {
				notifications.CloseNotification(notificationId)
			}
			previousPercentage = 101 // So when unplugging, and battery low, we get the relevant notification
		}
	}
}

var dbusConn = func() *dbus.Conn {
	if conn, err := dbus.SystemBus(); err != nil {
		panic(err)
	} else {
		return conn
	}
}()
