package icons

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/xdg"

	"github.com/surlykke/RefudeServices/lib/image"
)

func Run() {
	var baseDirs = []string{xdg.Home + "/.icons"}
	for _, dataDir := range xdg.DataDirs {
		baseDirs = append(baseDirs, dataDir+"/icons")
	}
	baseDirs = append(baseDirs, "/usr/share/pixmaps")

	initIconCollection(baseDirs)
	go recieveImages()
	go receiveIconDirs()

	fmt.Println("Adding basedirs", baseDirs)
	for _, baseDir := range baseDirs {
		baseDirSink <- baseDir
	}
	fmt.Println("Done")
}

func AddBaseDir(baseDir string) {
	baseDirSink <- baseDir
}

func AddARGBIcon(argbIcon image.ARGBIcon) string {
	var iconName = image.ARGBIconHashName(argbIcon)
	if reg.haveNotAdded(iconName) {
		for _, pixMap := range argbIcon.Images {
			if pixMap.Width != pixMap.Height {
			} else {
				var dir = fmt.Sprintf("%s/%d", refudeSessionIconsDir, pixMap.Width)
				go saveAsPng(dir, iconName, &pixMap)
				sessionIconImages <- sessionIconImage{
					name: iconName,
					size: pixMap.Width,
				}
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
			otherIconImages <- otherIconImage{name: name, path: filePath}
			return name
		}
	}
	return name
}

func AddRawImageIcon(imageData image.ImageData) string {
	var name = image.ImageDataHashName(imageData)
	if reg.haveNotAdded(name) {
		go saveAsPng(refudeSessionIconsDir, name, imageData)
		otherIconImages <- otherIconImage{
			name: name,
			path: refudeSessionIconsDir + "/" + name + ".png",
		}
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
