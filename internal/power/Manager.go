// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package power

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/lib/icon"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/notifications"
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

func retrieveDevice(dbusPath dbus.ObjectPath) (string, *Device) {

	var device = Device{Id: dbusPath2id(dbusPath)}
	device.DisplayDevice = dbusPath == displayDeviceDbusPath
	var props = utils.GetAllProps(dbusConn, upowerService, dbusPath, upowerDeviceInterface)

	device.NativePath, _ = props["NativePath"].Value().(string)
	device.Vendor, _ = props["Vendor"].Value().(string)
	device.Model, _ = props["Model"].Value().(string)
	device.Serial, _ = props["Serial"].Value().(string)
	device.UpdateTime, _ = props["UpdateTime"].Value().(uint64)
	var dType, _ = props["Type"].Value().(uint32)
	device.Type = deviceType(dType)
	device.PowerSupply, _ = props["PowerSupply"].Value().(bool)
	device.HasHistory, _ = props["HasHistory"].Value().(bool)
	device.HasStatistics, _ = props["HasStatistics"].Value().(bool)
	device.Online, _ = props["Online"].Value().(bool)
	device.Energy, _ = props["Energy"].Value().(float64)
	device.EnergyEmpty, _ = props["EnergyEmpty"].Value().(float64)
	device.EnergyFull, _ = props["EnergyFull"].Value().(float64)
	device.EnergyFullDesign, _ = props["EnergyFullDesign"].Value().(float64)
	device.EnergyRate, _ = props["EnergyRate"].Value().(float64)
	device.Voltage, _ = props["Voltage"].Value().(float64)
	device.Luminosity, _ = props["Luminosity"].Value().(float64)
	device.TimeToEmpty, _ = props["TimeToEmpty"].Value().(int64)
	device.TimeToFull, _ = props["TimeToFull"].Value().(int64)
	var percentage, _ = props["Percentage"].Value().(float64)
	device.Percentage = percentage
	device.IsPresent, _ = props["IsPresent"].Value().(bool)
	var state, _ = props["State"].Value().(uint32)
	device.State = deviceState(state)
	device.IsRechargeable, _ = props["IsRechargeable"].Value().(bool)
	device.Capacity, _ = props["Capacity"].Value().(float64)
	var tech, _ = props["Technology"].Value().(uint32)
	device.Technology = deviceTecnology(tech)
	var warnL, _ = props["WarningLevel"].Value().(uint32)
	device.Warninglevel = deviceWarningLevel(warnL)
	var batL, _ = props["BatteryLevel"].Value().(uint32)
	device.Batterylevel = deviceBatteryLevel(batL)
	var title = deviceTitle(device.Type, device.Model)
	var iconName, _ = props["IconName"].Value().(string)

	device.Base = *entity.MakeBase(title, "", icon.Name(iconName), entity.Device, "battery")
	return device.Id, &device
}

var previousPercentage = float64(101)

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
	if displayDevice, ok := DeviceMap.Get(dbusPath2id(displayDeviceDbusPath)); ok {
		if displayDevice.State == "Discharging" {
			if displayDevice.Percentage <= 5 {
				notifications.Notify("refude", notificationId, "dialog-warning", "Battery critical", fmt.Sprintf("At %.2f%%", displayDevice.Percentage), []string{}, map[string]dbus.Variant{"urgency": dbus.MakeVariant(uint8(2))}, -1)
			} else if displayDevice.Percentage <= 10 && previousPercentage > 10 {
				notifications.Notify("refude", notificationId, "dialog-information", "Battery", fmt.Sprintf("At %.2f%%", displayDevice.Percentage), []string{}, map[string]dbus.Variant{}, 10000)
			} else if displayDevice.Percentage <= 15 && previousPercentage > 15 {
				notifications.Notify("refude", notificationId, "dialog-information", "Battery", fmt.Sprintf("At %.2f%%", displayDevice.Percentage), []string{}, map[string]dbus.Variant{}, 5000)
			}
			previousPercentage = displayDevice.Percentage
		} else {
			notifications.CloseNotification(notificationId)
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
