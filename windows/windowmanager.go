package windows

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/windows/wayland"
	"github.com/surlykke/RefudeServices/windows/x11"
)

type WindowManager interface {
	Search(term string) link.List
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	GetPaths() []string
	RaiseAndFocusNamedWindow(name string) bool
	ResizeNamedWindow(name string, newWidth, newHeight uint32) bool
	HaveNamedWindow(name string) bool
	Run()
}

var WM WindowManager

func init() {
	if x11.WM != nil {
		WM = x11.WM 
	} else if wayland.WM != nil {
		WM = wayland.WM
	} else {
		panic("Ingen wm")
	}
}
