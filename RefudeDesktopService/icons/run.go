// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/respond"

	"github.com/surlykke/RefudeServices/lib/xdg"

	"github.com/surlykke/RefudeServices/lib/image"
)

func Handler(r *http.Request) http.Handler {
	if r.URL.Path == "/icon" {
		return IconResource{}
	} else if r.URL.Path == "/iconthemes" {
		return CollectThemes()
	} else if strings.HasPrefix(r.URL.Path, "/icontheme/") {
		if theme := getTheme(r.URL.Path); theme == nil {
			return nil
		} else {
			return theme
		}
	} else {
		return nil
	}
}

func CollectThemes() respond.StandardFormatList {
	lock.Lock()
	defer lock.Unlock()
	var sfl = make(respond.StandardFormatList, 0, len(themes))
	for _, theme := range themes {
		sfl = append(sfl, theme.ToStandardFormat())
	}

	return sfl
}

func AllPaths() []string {
	lock.Lock()
	defer lock.Unlock()
	var paths = make([]string, 0, len(themes)+1)
	for path, _ := range themes {
		paths = append(paths, path)
	}
	paths = append(paths, "/icon")
	return paths
}

func Run() {
	AddBaseDir(xdg.Home + "/.icons")
	AddBaseDir(xdg.DataHome + "/icons")
	for _, dataDir := range xdg.DataDirs {
		AddBaseDir(dataDir + "/icons")
	}
	AddBaseDir("/usr/share/pixmaps")
}

func getTheme(path string) *IconTheme {
	lock.Lock()
	defer lock.Unlock()
	return themes[path]
}

func AddARGBIcon(argbIcon image.ARGBIcon) string {
	var iconName = image.ARGBIconHashName(argbIcon)
	if reg.haveNotAdded(iconName) {
		lock.Lock()
		defer lock.Unlock()
		for _, pixMap := range argbIcon.Images {
			if pixMap.Width == pixMap.Height { // else ignore
				var dir = fmt.Sprintf("%s/%d", refudeSessionIconsDir, pixMap.Width)
				saveAsPng(dir, iconName, &pixMap)

				sessionIcons[iconName] = append(sessionIcons[iconName], IconImage{
					MinSize: pixMap.Width,
					MaxSize: pixMap.Width,
					Path:    fmt.Sprintf("%s/%s.png", dir, iconName),
				})

			}
		}
	}
	return iconName
}

func AddFileIcon(filePath string) string {
	filePath = path.Clean(filePath)
	var name = strings.Replace(filePath[1:len(filePath)-4], "/", ".", -1)
	if reg.haveNotAdded(name) {
		if fileInfo, err := os.Stat(filePath); err != nil {
			fmt.Println("error stat'ing:", filePath, err)
			return ""
		} else if !fileInfo.Mode().IsRegular() {
			fmt.Println("Not a regular file:", filePath)
			return ""
		} else if !(strings.HasSuffix(filePath, ".png") || strings.HasSuffix(filePath, ".svg")) {
			fmt.Println("Not an icon  file", filePath)
			return ""
		} else {
			AddOtherIcon(name, filePath)
			return name
		}
	}
	return name
}

func AddRawImageIcon(imageData image.ImageData) string {
	var name = image.ImageDataHashName(imageData)
	if reg.haveNotAdded(name) {
		saveAsPng(refudeSessionIconsDir, name, imageData)
		AddOtherIcon(name, fmt.Sprintf("%s/%s.png", refudeSessionIconsDir, name))
	}
	return name
}

type ConcurrentStringSet struct {
	sync.Mutex
	added map[string]bool
}

func (css *ConcurrentStringSet) haveNotAdded(val string) bool {
	css.Lock()
	defer css.Unlock()
	if css.added[val] {
		return false
	} else {
		css.added[val] = true
		return true
	}
}

var reg = &ConcurrentStringSet{added: make(map[string]bool)}
