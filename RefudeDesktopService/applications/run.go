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
	"sort"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"golang.org/x/sys/unix"
)

var resourceMap = resource.MakeResourceMap()
var Resources = resource.MakeServer(resourceMap)

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
		var collection = make(map[string]resource.Res, len(mtc)+len(apps)+4)
		var mtlist = make(resource.Selfielist, 0, len(mtc))
		for _, mt := range mtc {
			collection[mt.Self] = resource.MakeJsonResource(mt)
			mtlist = append(mtlist, mt)
		}
		sort.Sort(mtlist)
		collection["/mimetypes"] = resource.MakeJsonResource(mtlist)
		collection["/mimetypes/brief"] = resource.MakeJsonResource(mtlist.GetSelfs())

		var applist = make(resource.Selfielist, 0, len(apps))
		for _, app := range apps {
			collection[app.Self] = resource.MakeJsonResource(app)
			applist = append(applist, app)
		}
		sort.Sort(applist)
		collection["/applications"] = resource.MakeJsonResource(applist)
		collection["/applications/brief"] = resource.MakeJsonResource(applist.GetSelfs())

		resourceMap.ReplaceAll(collection)

		if _, err := unix.Read(fd, dummy); err != nil {
			panic(err)
		}
		fmt.Println("Recollect apps and mimes...")
	}
}
