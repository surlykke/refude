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

var PowerResources = resource.MakeGenericResourceCollection()

func getDevice(path string) (*Device, bool) {
	jsonRes, ok := PowerResources.Get(path).(resource.JsonResource)
	if !ok {
		return nil, false
	}
	device, ok := jsonRes.Data.(*Device)
	if !ok {
		return nil, false
	}
	return device, true
}

func Run() {
	var signals = subscribeToDeviceUpdates()

	for _, device := range getDevices() {
		fmt.Println("Setting device to", device.GetSelf())
		PowerResources.Set(device.Self, resource.MakeJsonResource(device))
	}

	var session = buildSessionResource()
	PowerResources.Set(session.Self, resource.MakeJsonResource(session))

	for signal := range signals {
		if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			var path = deviceSelf(signal.Path)
			if device, ok := getDevice(path); ok {
				var copy = *device
				// Brute force here, we update all, as I've seen some problems with getting out of sync after suspend..
				updateDevice(&copy, dbuscall.GetAllProps(dbusConn, UPowService, signal.Path, UPowerDeviceInterface))
				PowerResources.Set(copy.Self, resource.MakeJsonResource(&copy))

			}

			// TODO Handle device added/removed
			// (need hardware to test)
		}
	}
}
