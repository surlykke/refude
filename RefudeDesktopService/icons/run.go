package icons

import (
	"fmt"
	"path"
	"strings"

	"github.com/surlykke/RefudeServices/lib/image"
)

var Icons = MakeIconCollection()

var BasedirSink = make(chan string)

type namedIcon struct {
	name string
	icon interface{} // ARGBIcon, ImageData or pngfile path
}

var iconSink = make(chan namedIcon)

func Run() {
	Icons.collect()
	for {
		select {
		case basedir := <-BasedirSink:
			Icons.addBasedir(basedir)
		case nIcon := <-iconSink:
			switch icon := nIcon.icon.(type) {
			case image.ARGBIcon:
				Icons.addARGBIcon(nIcon.name, icon)
			case image.ImageData:
				Icons.addImageDataIcon(nIcon.name, icon)
			case string:
				Icons.addPngFileIcon(nIcon.name, icon)
			}
		}
	}
}

func AddPngFromARGB(argbIcon image.ARGBIcon) string {
	var name = image.ARGBIconHashName(argbIcon)
	iconSink <- namedIcon{name, argbIcon}
	return name
}

func AddPngFromFile(filePath string) string {
	if !strings.HasSuffix(filePath, ".png") {
		fmt.Println("Not a png icon-file:", filePath)
		return ""
	}
	filePath = path.Clean(filePath)
	var name = strings.Replace(filePath[1:len(filePath)-4], "/", ".", -1)
	iconSink <- namedIcon{name, filePath}
	return name
}

func AddPngFromRawImage(imageData image.ImageData) string {
	var name = image.ImageDataHashName(imageData)
	iconSink <- namedIcon{name, imageData}
	return name
}
