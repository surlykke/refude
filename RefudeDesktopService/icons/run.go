package icons

import (
	"net/http"
	"strings"
)

var iconCollection = MakeIconCollection()

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
	go collectAndMonitorThemeIcons()
	go collectAndMonitorOtherIcons()
}

