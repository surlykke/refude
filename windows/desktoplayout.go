package windows

import (
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/windows/x11"
)

type DesktopLayout struct {
	Geometry Bounds
	Monitors []*x11.MonitorData
}

func (dl *DesktopLayout) Links() link.List {
	return link.MakeList("/desktoplayout", "DesktopLayout", "")
}

func (dl *DesktopLayout) RefudeType() string {
	return "desktoplayout"
}
