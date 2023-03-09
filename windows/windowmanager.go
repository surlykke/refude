package windows

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/windows/monitor"
	"github.com/surlykke/RefudeServices/windows/wayland"
	"github.com/surlykke/RefudeServices/windows/x11"
)

type WindowManager interface {
	Search(sink chan link.Link, term string)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	GetPaths() []string
	RaiseAndFocusNamedWindow(name string) bool
	GetMonitors() []*monitor.MonitorData
	Run()
}

var WM WindowManager

func init() {
	if x11.WM != nil {
		WM = x11.WM 
	} else if wayland.WM != nil {
		WM = wayland.WM
	} else {
		panic("No wm")
	}
}
