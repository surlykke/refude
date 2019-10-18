// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package power

import (
	"log"

	dbuscall "github.com/surlykke/RefudeServices/lib/dbusutils"
	"github.com/surlykke/RefudeServices/lib/resource"
)

var devices = make(map[string]*Device)

func Run() {
	var session = buildSessionResource()
	resource.MapSingle(session.Self, session)

	var signals = subscribeToDeviceUpdates()

	for _, device := range getDevices() {
		devices[device.Self] = device
	}
	updateDeviceList()

	for signal := range signals {
		if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			var path = deviceSelf(signal.Path)
			if device, ok := devices[path]; ok {
				updateDevice(device, dbuscall.GetAllProps(dbusConn, UPowService, signal.Path, UPowerDeviceInterface))
				updateDeviceList()
			} else {
				log.Println("Warn: Update on unknown device: ", signal.Path)
			}
		}
	}
}

func updateDeviceList() {
	var collection = make(map[string]interface{})
	for _, device := range devices {
		collection[device.Self] = &(*device)
	}
	var devicePaths, deviceList = resource.ExtractPathAndResourceLists(collection)
	collection["/devicepaths"] = devicePaths
	collection["/devices"] = deviceList
	resource.MapCollection(&collection, "devices")
}
