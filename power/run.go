// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package power

import (
	"github.com/godbus/dbus/v5"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
)

func Run() {
	var signals = subscribe()
	
	resourcerepo.Put(retrieveDevice(displayDeviceDbusPath))
	showOnDesktop()

	for _, dbusPath := range retrieveDevicePaths() {
		resourcerepo.Put(retrieveDevice(dbusPath))
	}

	for signal := range signals {
		if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			resourcerepo.Update(retrieveDevice(signal.Path))
			if (displayDeviceDbusPath == signal.Path) {
				showOnDesktop()
			}
		} else if signal.Name == "org.freedesktop.UPower.DeviceAdded" {
			if path, ok := getAddedRemovedPath(signal); ok {
				log.Info("Adding device:", path)
				resourcerepo.Put(retrieveDevice(path))
			}
		} else if signal.Name == "org.freedesktop.UPower.DeviceRemoved" {
			if path, ok := signal.Body[0].(dbus.ObjectPath); ok {
				log.Info("Deleting device:", path)
				resourcerepo.Remove(path2id(path))
			}
		} else {
			log.Warn("Update on unknown device: ", signal.Path)
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
