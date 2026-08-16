// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package power

import (
	"fmt"
	"log"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/refude/internal/lib/resource"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/notifications"
)

const upowerService = "org.freedesktop.UPower"
const upowerPath = "/org/freedesktop/UPower"
const upowerInterface = "org.freedesktop.UPower"
const devicePrefix = "/org/freedesktop/UPower/devices/"
const displayDeviceDbusPath = dbus.ObjectPath(devicePrefix + "DisplayDevice")
const displayDeviceId = "DisplayDevice"
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
	if enumCall.Err != nil {
		log.Println("Error on call to upower:", enumCall.Err)
		return nil
	}
	return append(enumCall.Body[0].([]dbus.ObjectPath), displayDeviceDbusPath)
}

func retrieveDevice(dbusPath dbus.ObjectPath) Device {
	var device = Device{Id: dbusPath2id(dbusPath)}
	device.DisplayDevice = dbusPath == displayDeviceDbusPath
	var props, err = utils.Props(dbusConn, upowerService, dbusPath, upowerDeviceInterface)
	if err != nil {
		log.Println(err)
	}

	device.NativePath, _ = props["NativePath"].(string)
	device.Vendor, _ = props["Vendor"].(string)
	device.Model, _ = props["Model"].(string)
	device.Serial, _ = props["Serial"].(string)
	device.UpdateTime, _ = props["UpdateTime"].(uint64)
	var dType, _ = props["Type"].(uint32)
	device.Type = deviceType(dType)
	device.PowerSupply, _ = props["PowerSupply"].(bool)
	device.HasHistory, _ = props["HasHistory"].(bool)
	device.HasStatistics, _ = props["HasStatistics"].(bool)
	device.Online, _ = props["Online"].(bool)
	device.Energy, _ = props["Energy"].(float64)
	device.EnergyEmpty, _ = props["EnergyEmpty"].(float64)
	device.EnergyFull, _ = props["EnergyFull"].(float64)
	device.EnergyFullDesign, _ = props["EnergyFullDesign"].(float64)
	device.EnergyRate, _ = props["EnergyRate"].(float64)
	device.Voltage, _ = props["Voltage"].(float64)
	device.Luminosity, _ = props["Luminosity"].(float64)
	device.TimeToEmpty, _ = props["TimeToEmpty"].(int64)
	device.TimeToFull, _ = props["TimeToFull"].(int64)
	var percentage, _ = props["Percentage"].(float64)
	device.Percentage = percentage
	device.IsPresent, _ = props["IsPresent"].(bool)
	var state, _ = props["State"].(uint32)
	device.State = deviceState(state)
	device.IsRechargeable, _ = props["IsRechargeable"].(bool)
	device.Capacity, _ = props["Capacity"].(float64)
	var tech, _ = props["Technology"].(uint32)
	device.Technology = deviceTecnology(tech)
	var warnL, _ = props["WarningLevel"].(uint32)
	device.Warninglevel = deviceWarningLevel(warnL)
	var batL, _ = props["BatteryLevel"].(uint32)
	device.Batterylevel = deviceBatteryLevel(batL)
	var iconName, _ = props["IconName"].(string)

	var title string
	if device.PowerSupply {

		if device.DisplayDevice {
			title = "Laptop battery"
		} else if device.Type == "Battery" {
			title = "Laptop battery " + device.Model
		} else if device.Type == "Line Power" {
			title = "Power supply"
		} else {
			title = device.Model
		}

	} else {
		title = device.Model
	}

	var summary string
	if device.Type == "Line Power" {
		if device.Online {
			summary = "Connected"
		} else {
			summary = "Disconnected"
		}
	} else {
		summary = fmt.Sprintf("%.0f%%", device.Percentage)
		if device.State == "Charging" || device.State == "Discharging" {
			summary = summary + " " + strings.ToLower(device.State)
		}
	}

	device.Base = resource.MakeBase(devicePathPrefix+device.Id, title, summary, iconName, "battery")
	return device
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
	if displayDevice, ok := resource.Get[Device](devicePathPrefix + displayDeviceId); ok {
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
