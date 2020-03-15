// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package power

import (
	"log"
	"net/http"
	"sync"

	"github.com/surlykke/RefudeServices/lib/searchutils"

	"github.com/godbus/dbus"
	"github.com/surlykke/RefudeServices/lib/respond"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if device := getDevice(r.URL.Path); device != nil {
		if r.Method == "GET" {
			respond.AsJson(w, device.ToStandardFormat())
		} else {
			respond.NotAllowed(w)
		}
	}
}

func SearchDevices(collector *searchutils.Collector) {
	deviceLock.Lock()
	defer deviceLock.Unlock()

	for _, device := range devices {
		collector.Collect(device.ToStandardFormat())
	}

}

func AllPaths() []string {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	var paths = make([]string, 0, len(devices))
	for path, _ := range devices {
		paths = append(paths, path)
	}
	return paths
}

func Run() {
	var knownPaths = map[dbus.ObjectPath]bool{DisplayDevicePath: true}
	var signals = subscribeToDeviceUpdates()

	setDevice(retrieveDevice(DisplayDevicePath))
	for _, dbusPath := range retrieveDevicePaths() {
		setDevice(retrieveDevice(dbusPath))
		knownPaths[dbusPath] = true
	}

	for signal := range signals {
		if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			if knownPaths[signal.Path] {
			}
			setDevice(retrieveDevice(signal.Path))
		} else {
			log.Println("Warn: Update on unknown device: ", signal.Path)
		}
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
	devices[deviceSelf(device)] = device
}
