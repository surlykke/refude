package rawfile

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/surlykke/refude/internal/lib/respond"
)

func Run() {
	http.HandleFunc(rawfilePathPrefix+"/", RawFileHandler)
}

const rawfilePathPrefix = "/rawfile"

var rawFiles = make(map[string]bool, 20)
var rawFileLock sync.Mutex

func RegisterRawFile(osPath string) string {
	osPath = filepath.Clean(osPath)
	var path = rawfilePathPrefix + osPath
	rawFileLock.Lock()
	defer rawFileLock.Unlock()
	rawFiles[path] = true
	return path
}

func isRegistered(path string) bool {
	rawFileLock.Lock()
	defer rawFileLock.Unlock()
	return rawFiles[path]
}

func RawFileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("rawfile:", r.URL.EscapedPath())
	if !isRegistered(r.URL.EscapedPath()) {
		respond.NotFound(w)
	} else {
		var filePath = r.URL.EscapedPath()[len(rawfilePathPrefix):]
		fmt.Println("Serving:", filePath)
		http.ServeFile(w, r, filePath)
	}
}
