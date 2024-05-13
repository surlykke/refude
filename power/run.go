// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package power

import (
	"github.com/godbus/dbus/v5"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/repo"
)

var deviceRepo = repo.MakeRepoWithFilter(searchFilter)
var requests = repo.MakeAndRegisterRequestChan()

var updates = make(chan *Device)
var removals = make(chan string)

func Run() {
	go dbusLoop()

	for {
		select {
		case req := <-requests:
			deviceRepo.DoRequest(req)
		case device := <-updates:
			deviceRepo.Put(device)
			if device.Path == "/device/DisplayDevice" {
				showOnDesktop()
			}
		case path := <-removals:
			deviceRepo.Remove(path)
		}
	}
}

func dbusLoop() {
	var signals = subscribe()

	updates <- retrieveDevice(displayDeviceDbusPath)

	for _, dbusPath := range retrieveDevicePaths() {
		updates <- retrieveDevice(dbusPath)
	}

	for signal := range signals {
		if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			updates <- retrieveDevice(signal.Path)
		} else if signal.Name == "org.freedesktop.UPower.DeviceAdded" {
			if path, ok := getAddedRemovedPath(signal); ok {
				updates <- retrieveDevice(path)
			}
		} else if signal.Name == "org.freedesktop.UPower.DeviceRemoved" {
			if path, ok := signal.Body[0].(dbus.ObjectPath); ok {
				removals <- "/device/" + path2id(path)
			}
		} else {
			log.Warn("Update on unknown device: ", signal.Path)
		}
	}

}

func searchFilter(term string, device *Device) bool {
	return len(term) > 2
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
