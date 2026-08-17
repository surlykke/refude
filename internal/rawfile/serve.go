// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
//
package rawfile

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/surlykke/refude/internal/lib/respond"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/lib/xdg"
)

func Run() {
	http.HandleFunc(rawfilePathPrefix+"/", RawFileHandler)
}

const rawfilePathPrefix = "/rawfile"
const cacheHeader = "max-age=31536000"
func RawFileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("rawfile:", r.URL.EscapedPath())
	if !get(r.URL.EscapedPath()) {
		respond.NotFound(w)
	} else {
		w.Header().Add("Cache-Control", cacheHeader)
		var filePath = r.URL.EscapedPath()[len(rawfilePathPrefix):]
		fmt.Println("Serving:", filePath)
		http.ServeFile(w, r, filePath)
	}
}

func RegisterRawFile(osPath string) string {
	osPath = filepath.Clean(osPath)
	var path = rawfilePathPrefix + osPath
	put(path)
	return path
}

func InstallAndRegisterBytes(bytes []byte, fileType string) string {
	var osPath = xdg.RefudeCacheDir + "/" + utils.HashName(bytes) + "." + fileType
	var path = rawfilePathPrefix + osPath
	if !get(path) {
		if err := os.WriteFile(osPath, bytes, 0600); err != nil {
			log.Print("Error saving", osPath, err)
			return ""
		}
	}
	put(path)
	return path

}

var reg = make(map[string]bool, 20)
var regLock sync.Mutex

func put(path string) {
	regLock.Lock()
	defer regLock.Unlock()
	reg[path] = true
}

func get(path string) bool {
	regLock.Lock()
	defer regLock.Unlock()
	return reg[path]
}
