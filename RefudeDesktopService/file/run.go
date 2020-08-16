package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func init() {
	if err := magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR); err != nil {
		panic(err)
	}
}

func Run() {
	if watcher, err := fsnotify.NewWatcher(); err != nil {
		fmt.Println(err)
	} else {
		watcher.Add(xdg.DownloadDir)
		for ev := range watcher.Events {
			if ev.Op&fsnotify.Create == fsnotify.Create && worthyOfAttention(ev.Name) {
				if file, err := makeFile(ev.Name); err == nil {
					addRecentDownload(file.Path)
					/*var fileName = filepath.Base(file.Path)
					var iconName string = "folder-download"
					if file.Mimetype != "" {
						iconName = strings.ReplaceAll(file.Mimetype, "/", "-")
					}
					//FIXME osd.PublishMessage(0, "org.refude.RefudeServices", "New download", fileName, iconName)*/

				}
			}
		}
	}
}

func worthyOfAttention(path string) bool {
	if !strings.HasPrefix(path, xdg.DownloadDir) {
		return false
	} else if strings.HasPrefix(filepath.Base(path), ".") ||
		strings.HasSuffix(path, ".part") || // Firefox partial download
		strings.HasSuffix(path, "crdownload") { // Chrome partial download

		return false
	} else if fileInfo, err := os.Stat(path); err != nil {
		fmt.Println("Error stat'ing", path, err)
		return false
	} else if fileInfo.Size() == 0 {
		return false
	} else {
		return true
	}
}
