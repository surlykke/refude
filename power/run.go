// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package power

import (
	"github.com/godbus/dbus/v5"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/watch"
)

func Run() {
	var signals = subscribe()

	setDevice(retrieveDevice(displayDeviceDbusPath))
	for _, dbusPath := range retrieveDevicePaths() {
		setDevice(retrieveDevice(dbusPath))
	}

	for signal := range signals {
		var path = devicePath(signal.Path)
		if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			if res := Devices.Get(path); res != nil {
				setDevice(retrieveDevice(signal.Path))
			}
		} else if signal.Name == "org.freedesktop.UPower.DeviceAdded" {
			if path, ok := getAddedRemovedPath(signal); ok {
				log.Info("Adding device:", path)
				setDevice(retrieveDevice(path))
			}
		} else if signal.Name == "org.freedesktop.UPower.DeviceRemoved" {
			if path, ok := signal.Body[0].(dbus.ObjectPath); ok {
				log.Info("Deleting device:", path)
				Devices.Delete(devicePath(path))
			}
		} else {
			log.Warn("Update on unknown device: ", signal.Path)
		}
		if path == displayDevicePath {
			watch.SomethingChanged(displayDevicePath)
		}
	}

}

func getAddedRemovedPath(signal *dbus.Signal) (dbus.ObjectPath, bool) {
	if len(signal.Body) == 0 {
		return "", false
	} else if path, ok := signal.Body[0].(dbus.ObjectPath); !ok {
		return "", false
	} else {
		return path, true
	}
}

var Devices = resource.MakeList("/device/list")

func setDevice(dev *Device) {
	Devices.Put(resource.MakeResource(devicePath(dev.DbusPath), dev.title, "", dev.IconName, "device", dev))
}
