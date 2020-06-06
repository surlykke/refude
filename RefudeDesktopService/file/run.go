package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/godbus/dbus/v5"
	"github.com/rakyll/magicmime"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/notifications"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func init() {
	if err := magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR); err != nil {
		panic(err)
	}
}

func Run() {
	var noActions = []string{}
	var noHints = map[string]dbus.Variant{}
	if watcher, err := fsnotify.NewWatcher(); err != nil {
		fmt.Println(err)
	} else {
		watcher.Add(xdg.DownloadDir)
		for ev := range watcher.Events {
			if ev.Op&fsnotify.Create == fsnotify.Create && worthyOfAttention(ev.Name) {
				var fileName = filepath.Base(ev.Name)
				notifications.Notify("RefudeServices", 0, "folder-download", "New download", fileName, noActions, noHints, 20000)
			}
		}
	}
}

func worthyOfAttention(path string) bool {
	if !strings.HasPrefix(path, xdg.DownloadDir) {
		return false
	} else if strings.HasPrefix(filepath.Base(path), ".") ||
		strings.HasSuffix(path, ".part" /**/) ||
		strings.HasSuffix(path, "crdownload") {

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
