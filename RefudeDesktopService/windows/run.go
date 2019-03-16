package windows

import (
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows/xlib"
	"net/http"
	"strings"
)

var Windows = MakeWindowCollection()

func Serve(w http.ResponseWriter, r *http.Request) bool {
	if strings.HasPrefix(r.URL.Path, "/window") {
		if r.Method == "GET" {
			Windows.GET(w, r)
		} else if r.Method == "POST" {
			Windows.POST(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return true
	} else {
		return false
	}
}

func Run() {
	var manager = Manager{
		in:  xlib.MakeConnection(),
		out: xlib.MakeConnection(),
	}
	manager.Run()
}
