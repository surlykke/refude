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

var ApplicationsAndMimetypes = func() *resource.GenericResourceCollection {
	var grc = resource.MakeGenericResourceCollection()
	grc.AddCollectionResource("/applications", "/application/")
	grc.AddCollectionResource("/mimetypes", "/mimetype/")
	return grc
}()

func Run() {
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

	if _, err := unix.InotifyAddWatch(fd, xdg.ConfigHome+"/mimeapps.list", unix.IN_CLOSE_WRITE); err != nil {
		panic(err)
	}

	dummy := make([]byte, 100)
	for {
		var mtc, apps = Collect()
		var collection = make(map[string]resource.Resource, len(mtc)+len(apps))
		for _, mt := range mtc {
			collection[string(mt.GetSelf())] = mt
		}
		for _, app := range apps {
			collection[string(app.GetSelf())] = app
		}
		ApplicationsAndMimetypes.ReplaceAll(collection)

		if _, err := unix.Read(fd, dummy); err != nil {
			panic(err)
		}
		fmt.Println("applications do new collect...")
	}
}
