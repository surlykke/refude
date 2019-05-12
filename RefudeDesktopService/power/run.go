// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package power

import (
	"fmt"

	dbuscall "github.com/surlykke/RefudeServices/lib/dbusutils"
	"github.com/surlykke/RefudeServices/lib/resource"
)

var powerMap = resource.MakeResourceMap("/devices")
var PowerResources = resource.MakeJsonResourceServer(powerMap)

func getDevice(path string) (*Device, bool) {
	device, ok := PowerResources.Get(path).(*Device)
	return device, ok
}

func Run() {
	var signals = subscribeToDeviceUpdates()

	for _, device := range getDevices() {
		fmt.Println("Setting device to", device.Self)
		powerMap.Set(device.Self, device)
	}

	var session = buildSessionResource()
	powerMap.Set(session.Self, session)

	for signal := range signals {
		if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			var path = deviceSelf(signal.Path)
			if device, ok := getDevice(path); ok {
				var copy = *device
				// Brute force here, we update all, as I've seen some problems with getting out of sync after suspend..
				updateDevice(&copy, dbuscall.GetAllProps(dbusConn, UPowService, signal.Path, UPowerDeviceInterface))
				powerMap.Set(copy.Self, &copy)

			}

			// TODO Handle device added/removed
			// (need hardware to test)
		}
	}
}
