package windows

import (
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/wayland"
	"github.com/surlykke/RefudeServices/x11"
)

var runningX11 = xdg.SessionType == "x11"

func PurgeAndShow(applicationTitle string, focus bool) bool {
	if runningX11 {
		return x11.PurgeAndShow(applicationTitle, focus) 
	} else {
		return wayland.PurgeAndShow(applicationTitle, focus)
	}
}

