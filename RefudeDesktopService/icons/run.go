package icons

import (
	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

var basedirSink = make(chan string)
var iconSink = make(chan image.ARGBIcon)

func AddBasedir(dirToLookAt string) {
	basedirSink <- dirToLookAt
}

func AddARGBIcon(icon image.ARGBIcon) {
	iconSink <- icon
}

func Run() {
	addBaseDir(xdg.Home + "/.icons")
	for _, dataDir := range xdg.DataDirs {
		addBaseDir(dataDir + "/icons")
	}
	addBaseDir(xdg.Home + "/.local/share/icons")
	addBaseDir(refudeSessionIconsDir)

	go monitorBasedirSink()

	// FIXME iconSink

}
