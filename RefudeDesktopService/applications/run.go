package applications

import (
	"fmt"
	"os"

	"github.com/surlykke/RefudeServices/lib/xdg"
	"golang.org/x/sys/unix"
)

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
		mlock.Lock()
		mimetypes = mtc
		mlock.Unlock()

		lock.Lock()
		desktopApplications = apps
		lock.Unlock()

		if _, err := unix.Read(fd, dummy); err != nil {
			panic(err)
		}
	}
}
