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
)

func Run() {
	fmt.Println("power.Run")
	var signals = setup()
	fmt.Println("looking for signals")

	for signal := range signals {
		//fmt.Println("Signal: ", signal)
		if signal.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			var path = deviceSelf(signal.Path)
			if device := GetDevice(path); device != nil {
				var copy = *device
				// Brute force here, we update all, as I've seen some problems with getting out of sync after suspend..
				updateDevice(&copy, dbuscall.GetAllProps(dbusConn, UPowService, signal.Path, UPowerDeviceInterface))
				setDevice(&copy)

			}
			// TODO Handle device added/removed
			// (need hardware to test)
		}
	}
}
