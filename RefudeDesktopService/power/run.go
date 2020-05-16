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

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/watch"
	"github.com/surlykke/RefudeServices/lib/searchutils"

	"github.com/surlykke/RefudeServices/lib/respond"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/devices" {
		respond.AsJson(w, r, Collect(searchutils.Term(r)))
	} else if device := getDevice(r.URL.Path); device != nil {
		respond.AsJson(w, r, device.ToStandardFormat())
	}
}

func Collect(term string) respond.StandardFormatList {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	var sfl = make(respond.StandardFormatList, 0, len(devices))
	for _, device := range devices {
		if rank := searchutils.SimpleRank(string(device.DbusPath), "", term); rank > -1 {
			sfl = append(sfl, device.ToStandardFormat().Ranked(rank))
		}
	}

	return sfl.SortByRank()
}

func AllPaths() []string {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	var paths = make([]string, 0, len(devices)+1)
	for path, _ := range devices {
		paths = append(paths, path)
	}
	paths = append(paths, "/devices")
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
	var self = deviceSelf(device)
	devices[self] = device
	watch.SomethingChanged(self)
}
