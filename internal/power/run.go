// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package power

import (
	"log"

	"github.com/godbus/dbus/v5"

	"github.com/surlykke/refude/internal/lib/resource"
	"github.com/surlykke/refude/pkg/pubsub"
)

const devicePathPrefix = "/device/"
var deviceEvents = pubsub.MakePublisher[struct{}]()

func Run(dontShowTrayBattery bool) {
	var signals = subscribe()

	resource.Put(retrieveDevice(displayDeviceDbusPath))
	showOnDesktop()

	for _, dbusPath := range retrieveDevicePaths() {
		resource.Put(retrieveDevice(dbusPath))
	}
	if !dontShowTrayBattery {
		go tray_applet_run()
	}

	for signal := range signals {
		switch signal.Name {
		case "org.freedesktop.DBus.Properties.PropertiesChanged":
			var device = retrieveDevice(signal.Path)
			resource.Put(device)
			if device.DisplayDevice {
				showOnDesktop()
			}
		case "org.freedesktop.UPower.DeviceAdded":
			if path, ok := getAddedRemovedPath(signal); ok {
				resource.Put(retrieveDevice(path))
			}
		case "org.freedesktop.UPower.DeviceRemoved":
			if dbusPath, ok := signal.Body[0].(dbus.ObjectPath); ok {
				resource.Remove(devicePathPrefix + dbusPath2id(dbusPath))
			}
		default:
			log.Print("Update on unknown device: ", signal.Path)
		}
		deviceEvents.Publish(struct{}{})
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
