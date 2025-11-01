// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package power

import (
	"log"

	"github.com/godbus/dbus/v5"

	"github.com/surlykke/refude/internal/lib/entity"
)

var DeviceMap = entity.MakeMap[string, *Device]()

func Run(dontShowTrayBattery bool) {
	var signals = subscribe()

	DeviceMap.Put(retrieveDevice(displayDeviceDbusPath))
	showOnDesktop()

	for _, dbusPath := range retrieveDevicePaths() {
		DeviceMap.Put(retrieveDevice(dbusPath))
	}
	if !dontShowTrayBattery {
		go tray_applet_run()
	}

	for signal := range signals {
		switch signal.Name {
		case "org.freedesktop.DBus.Properties.PropertiesChanged":
			var id, device = retrieveDevice(signal.Path)
			DeviceMap.Put(id, device)
			if device.DisplayDevice {
				showOnDesktop()
			}
		case "org.freedesktop.UPower.DeviceAdded":
			if path, ok := getAddedRemovedPath(signal); ok {
				DeviceMap.Put(retrieveDevice(path))
			}
		case "org.freedesktop.UPower.DeviceRemoved":
			if dbusPath, ok := signal.Body[0].(dbus.ObjectPath); ok {
				DeviceMap.Remove(dbusPath2id(dbusPath))
			}
		default:
			log.Print("Update on unknown device: ", signal.Path)
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
