// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"regexp"

	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

var BasedirSink = make(chan string)
var IconSink = make(chan image.ARGBIcon)
var IconRepo = resource.MakeGenericResourceCollection()

func Run() {
	IconRepo.Set("/icons", IconRepo.MakeRegexpCollection(regexp.MustCompile(`^/icon/[^/]+$`)))
	addBaseDir(xdg.Home + "/.icons")
	addBaseDir(xdg.Home + "/.local/share/icons") // Not in the icon theme spec, but I think it should be
	for _, dataDir := range xdg.DataDirs {
		addBaseDir(dataDir + "/icons")
	}
	addBaseDir("/usr/share/pixmaps")
	addBaseDir(refudeSessionIconsDir)

	for {
		publishFoundIcons()
		select {
		case baseDir := <-BasedirSink:
			addBaseDir(baseDir)
		case argbIcon := <-IconSink:
			addARGBIcon(argbIcon)
		}
	}
}
