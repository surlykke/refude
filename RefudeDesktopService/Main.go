// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/surlykke/RefudeServices/RefudeDesktopService/applications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/power"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows"
	"github.com/surlykke/RefudeServices/lib"
	"github.com/surlykke/RefudeServices/lib/resource"
)


func main() {
	var updateStream = make(chan resource.Update)
	var resourceMap = resource.MakeJsonResourceMap()

	go applications.Run(updateStream)
	go windows.Run(updateStream)
	go power.Run(updateStream);

	go resourceMap.Run(updateStream)

	lib.Serve("org.refude.desktop-service", resourceMap)
}
