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

var PowerResources = func() *resource.GenericResourceCollection {
	var grc = resource.MakeGenericResourceCollection()
	grc.AddCollectionResource("/devices", "/device/")
	return grc
}()

func Run() {
	var signals = subscribeToDeviceUpdates()

	for _, device := range getDevices() {
		fmt.Println("Setting device to", device.GetSelf())
		PowerResources.Set(device)
	}

	PowerResources.Set(buildSessionResource())

	for signal := range signals {
		if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			var path = deviceSelf(signal.Path)
			if res := PowerResources.Get(string(path)); res != nil {
				if device, ok := res.(*Device); ok {
					var copy = *device
					// Brute force here, we update all, as I've seen some problems with getting out of sync after suspend..
					updateDevice(&copy, dbuscall.GetAllProps(dbusConn, UPowService, signal.Path, UPowerDeviceInterface))
					PowerResources.Set(&copy)

				}
			}

			// TODO Handle device added/removed
			// (need hardware to test)
		}
	}
}
