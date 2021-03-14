package windows

import (
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/windows/x11"
)

type DesktopLayout struct {
	respond.Resource
	Geometry Bounds
	Monitors []*x11.MonitorData
}
