package windows

import (
	"math"
	"sync"

	"github.com/surlykke/RefudeServices/windows/x11"
)

type DesktopLayout struct {
	Geometry Bounds
	Monitors []*x11.MonitorData
}

var desktopLayout *DesktopLayout
var dlLock sync.Mutex

func getDesktopLayout() *DesktopLayout {
	dlLock.Lock()
	defer dlLock.Unlock()

	return desktopLayout
}

func setDesktopLayout(newDesktopLayout *DesktopLayout) {
	dlLock.Lock()
	defer dlLock.Unlock()
	desktopLayout = newDesktopLayout
}

func updateDesktopLayout(p x11.Proxy) {
	var monitors = x11.GetMonitorDataList(p)
	var layout = &DesktopLayout{
		Monitors: monitors,
	}

	var minX, minY = int32(math.MaxInt32), int32(math.MaxInt32)
	var maxX, maxY = int32(math.MinInt32), int32(math.MinInt32)

	for _, m := range layout.Monitors {
		if minX > m.X {
			minX = m.X
		}
		if minY > m.Y {
			minY = m.Y
		}

		if maxX < m.X+int32(m.W) {
			maxX = m.X + int32(m.W)
		}

		if maxY < m.Y+int32(m.H) {
			maxY = m.Y + int32(m.H)
		}
	}

	layout.Geometry = Bounds{minX, minY, uint32(maxX - minY), uint32(maxY - minY)}

	desktopLayout = layout
}
