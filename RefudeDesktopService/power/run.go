// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package power

import (
	"fmt"
	"sort"

	dbuscall "github.com/surlykke/RefudeServices/lib/dbusutils"
	"github.com/surlykke/RefudeServices/lib/resource"
)

var devices = make(map[string]*Device)
var powerMap = resource.MakeResourceMap()
var PowerResources = resource.MakeServer(powerMap)

func Run() {
	var signals = subscribeToDeviceUpdates()

	for _, device := range getDevices() {
		devices[device.Self] = device
		powerMap.Set(device.Self, resource.MakeJsonResouceWithEtag(device))
	}

	var session = buildSessionResource()
	powerMap.Set(session.Self, resource.MakeJsonResource(session))

	for signal := range signals {
		if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			var path = deviceSelf(signal.Path)
			if device, ok := devices[path]; ok {
				var copy = *device
				// Brute force here, we update all, as I've seen some problems with getting out of sync after suspend..
				updateDevice(&copy, dbuscall.GetAllProps(dbusConn, UPowService, signal.Path, UPowerDeviceInterface))
				powerMap.Set(copy.Self, resource.MakeJsonResouceWithEtag(&copy))
				powerMap.Broadcast()
			} else {
				fmt.Println("Update on unknown device: ", signal.Path)
			}

			// TODO Handle device added/removed
			// (need hardware to test)
		}
	}
}

func updateCollections() {
	var list = make(resource.Selfielist, 0, len(devices))
	for _, device := range devices {
		list = append(list, device)
	}
	sort.Sort(list)
	powerMap.Set("/devices", resource.MakeJsonResource(list))
	powerMap.Set("/devices/brief", resource.MakeJsonResource(list.GetSelfs()))
}
