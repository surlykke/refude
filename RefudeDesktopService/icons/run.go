package icons

import (
	"github.com/surlykke/RefudeServices/lib/xdg"
	"net/http"
	"strings"
)


var DirsToLookAt chan string


func Serve(w http.ResponseWriter, r *http.Request) bool {
	if ! strings.HasPrefix(r.URL.Path, "/icon") {
		return false
	}

	if r.Method == "GET" {
		iconCollection.GET(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	return true
}

func Run() {
	go consumeThemes()
	go consumeIcons()
	addBaseDir(xdg.Home + "/.icons")
	addBaseDir(xdg.DataHome + "/icons")
	for i := len(xdg.DataDirs) - 1; i >= 0; i-- {
		addBaseDir(xdg.DataDirs[i]+"/icons")
	}

	for dirToLookAt := range DirsToLookAt {
		addBaseDir(dirToLookAt)
	}
}

