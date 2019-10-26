// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package applications

import (
	"fmt"
	"os"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"golang.org/x/sys/unix"
)

func Run() {
	fmt.Println("Ind i applications.Run")
	fd, err := unix.InotifyInit()

	if err != nil {
		panic(err)
	}

	defer unix.Close(fd)

	for _, dataDir := range append(xdg.DataDirs, xdg.DataHome) {
		appDir := dataDir + "/applications"
		if _, err := os.Stat(appDir); os.IsNotExist(err) {
			// path/to/whatever does not exist
		}

		if xdg.DirOrFileExists(appDir) {
			if _, err := unix.InotifyAddWatch(fd, appDir, unix.IN_CREATE|unix.IN_MODIFY|unix.IN_DELETE); err != nil {
				fmt.Println("Could not watch:", appDir, ":", err)
			}
		}
	}

	// Make sure we have ~/.config/mimeapps.list	
	if _, err := os.Stat(xdg.ConfigHome + "/mimeapps.list"); err != nil && os.IsNotExist(err) {
		if emptyMimemappsList, err := os.Create(xdg.ConfigHome + "/mimeapps.list"); err != nil {
			panic(err)
		} else {
			emptyMimemappsList.Close()
		}
	}

	if _, err := unix.InotifyAddWatch(fd, xdg.ConfigHome+"/mimeapps.list", unix.IN_CLOSE_WRITE); err != nil {
		panic(err)
	}

	dummy := make([]byte, 100)
	for {
		fmt.Println("collect mimetypes and applicatons")
		var mimetypeResources, applicationResources = Collect()

		fmt.Println("found", len(mimetypeResources), "mimetypes and", len(applicationResources), "applications")

		resource.MapCollection(&mimetypeResources, "mimetypes")
		resource.MapCollection(&applicationResources, "applications")

		if _, err := unix.Read(fd, dummy); err != nil {
			panic(err)
		}
		fmt.Println("Recollect apps and mimes...")
	}
}
