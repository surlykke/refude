package windows

import (
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/windows/x11"
)

type DesktopLayout struct {
	Geometry Bounds
	Monitors []*x11.MonitorData
}

func (dl *DesktopLayout) Links() []resource.Link {
	return []resource.Link{resource.MakeLink("/desktoplayout", "DesktopLayout", "", relation.Self)}
}

func (dl *DesktopLayout) RefudeType() string {
	return "desktoplayout"
}
