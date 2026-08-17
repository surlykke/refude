// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package rawfile

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/surlykke/refude/internal/lib/respond"
)

const Prefix = "/rawfile"
const cacheHeader = "max-age=31536000"

func Run() {
	http.HandleFunc(Prefix+"/", RawFileHandler)
}

func RawFileHandler(w http.ResponseWriter, r *http.Request) {
	var osPath = r.URL.Path[len(Prefix):]
	fmt.Println("rawfile:", osPath)
	if !isRegistered(osPath) {
		respond.NotFound(w)
	} else {
		w.Header().Add("Cache-Control", cacheHeader)
		http.ServeFile(w, r, osPath)
	}
}

func Register(osPath string) string {
	regLock.Lock()
	defer regLock.Unlock()
	reg[osPath] = true
	return Prefix + osPath
}

func isRegistered(path string) bool {
	regLock.Lock()
	defer regLock.Unlock()
	return reg[path]
}

var reg = make(map[string]bool, 20)
var regLock sync.Mutex



