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

	var device = Device{}

	var props = dbuscall.GetAllProps(dbusConn, upowerService, dbusPath, upowerDeviceInterface)

	device.NativePath, _ = props["NativePath"].Value().(string)
	device.Vendor, _ = props["Vendor"].Value().(string)
	device.Model, _ = props["Model"].Value().(string)
	device.Serial, _ = props["Serial"].Value().(string)
	device.UpdateTime, _ = props["UpdateTime"].Value().(uint64)
	device.Type, _ = props["Type"].Value().(string)
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
	device.Percentage = int8(100 * percentage)
	device.IsPresent, _ = props["IsPresent"].Value().(bool)
	device.State, _ = props["State"].Value().(string)
	device.IsRechargeable, _ = props["IsRechargeable"].Value().(bool)
	device.Capacity, _ = props["Capacity"].Value().(float64)
	var tech, _ = props["Technology"].Value().(uint32)
	device.Technology = deviceTecnology(tech)
	var warnL, _ = props["WarningLevel"].Value().(uint32)
	device.Warninglevel = deviceWarningLevel(warnL)
	var batL, _ = props["BatteryLevel"].Value().(uint32)
	device.Batterylevel = deviceBatteryLevel(batL)
	device.Keywords = []string{"battery"}

	var title = deviceTitle(device.Type, device.Model)
	var iconName, _ = props["IconName"].Value().(string)

	device.ResourceData = *resource.MakeBase(path.Of("/device/", dbusPath2id(dbusPath)), title, "", icon.Name(iconName), mediatype.Device)

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
	if displayDevice, ok := repo.Get[*Device](path.Of("/device/", dbusPath2id(displayDeviceDbusPath))); ok {
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
