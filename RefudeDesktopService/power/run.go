// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package power

import (
	"fmt"
	"log"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/watch"

	"github.com/surlykke/RefudeServices/lib/respond"
)

func Collect() respond.StandardFormatList {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	var sfl = make(respond.StandardFormatList, 0, len(devices))
	for _, device := range devices {
		sfl = append(sfl, device.ToStandardFormat())
	}

	return sfl
}

func AllPaths() []string {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	var paths = make([]string, 0, len(devices)+1)
	for path := range devices {
		paths = append(paths, path)
	}
	paths = append(paths, "/devices")
	return paths
}

func Run() {
	var knownPaths = map[dbus.ObjectPath]bool{DisplayDevicePath: true}
	var signals = subscribe()

	setDevice(retrieveDevice(DisplayDevicePath))
	for _, dbusPath := range retrieveDevicePaths() {
		fmt.Println("Adding device", dbusPath)
		setDevice(retrieveDevice(dbusPath))
		knownPaths[dbusPath] = true
	}

	for signal := range signals {
		if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			if knownPaths[signal.Path] {
				setDevice(retrieveDevice(signal.Path))
			}
		} else if signal.Name == "org.freedesktop.UPower.DeviceAdded" {
			if path, ok := getAddedRemovedPath(signal); ok {
				fmt.Println("Adding device", path)
				setDevice(retrieveDevice(path))
			}
		} else if signal.Name == "org.freedesktop.UPower.DeviceRemoved" {
			if path, ok := signal.Body[0].(dbus.ObjectPath); ok {
				fmt.Println("Removing device", path)
				removeDevice(path)
			}
		} else {
			log.Println("Warn: Update on unknown device: ", signal.Path)
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

var devices = make(map[string]*Device)
var deviceLock sync.Mutex

func getDevice(path string) *Device {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	return devices[path]
}

func setDevice(device *Device) {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	var self = deviceSelf(device.DbusPath)
	devices[self] = device
	watch.SomethingChanged(self)
}

func removeDevice(path dbus.ObjectPath) {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	delete(devices, deviceSelf(path))
}
