package applications

import (
	"fmt"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"golang.org/x/sys/unix"
	"net/http"
	"os"
	"strings"
)

var applicationCollection = MakeDesktopApplicationCollection()
var mimetypeCollection = MakeMimetypecollection()

func Serve(w http.ResponseWriter, r *http.Request) bool {
	if strings.HasPrefix(r.URL.Path, "/application") {
		if r.Method == "GET" {
			applicationCollection.GET(w, r)
		} else if r.Method == "POST" {
			applicationCollection.POST(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return true
	} else if strings.HasPrefix(r.URL.Path, "/mimetype") {
		if r.Method == "GET" {
			mimetypeCollection.GET(w, r)
		} else if r.Method == "PATCH" {
			mimetypeCollection.PATCH(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return true
	}

	return false
}

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
			fmt.Println("Watching: " + appDir)
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
		var mtc, apps = Collect();
		mimetypeCollection.mutex.Lock()
		mimetypeCollection.mimetypes = mtc
		mimetypeCollection.CachingJsonGetter.Clear()
		mimetypeCollection.mutex.Unlock()

		applicationCollection.mutex.Lock()
		applicationCollection.apps = apps
		applicationCollection.CachingJsonGetter.Clear()
		applicationCollection.mutex.Unlock()

		if _, err := unix.Read(fd, dummy); err != nil {
			panic(err)
		}
	}
}
