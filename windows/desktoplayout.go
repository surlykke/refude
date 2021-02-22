package windows

import (
	"math"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/respond"
)

type DesktopLayout struct {
	respond.Links
	Geometry Bounds
	Monitors []*MonitorData
}

func getDesktopLayout(c *Connection, oldWindowList []*Window) (*DesktopLayout, []*Window) {
	var monitors = c.GetMonitorDataList()
	var layout = &DesktopLayout{
		Links:    respond.Links{respond.Link{Href: "/desktoplayout", Title: "DesktopLayout", Rel: respond.Self, Profile: "/profile/desktoplayout"}},
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

	var newWindowList = make([]*Window, len(oldWindowList), len(oldWindowList))
	for i, win := range oldWindowList {
		var copy = *win
		updateLinksSingle(&copy, monitors)
		newWindowList[i] = &copy
	}

	return layout, newWindowList
}

func (dl *DesktopLayout) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, dl)
	} else {
		respond.NotAllowed(w)
	}
}
