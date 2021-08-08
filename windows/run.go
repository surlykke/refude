// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"math"
	"strconv"
	"sync"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/windows/x11"
)

// Maintains windows  and monitors lists
func Run() {
	var proxy = x11.MakeProxy()
	x11.SubscribeToEvents(proxy)

	updateDesktopLayout(proxy)
	updateWindowList(proxy)

	for {
		var change = false

		if event, wId := x11.NextEvent(proxy); event == x11.DesktopGeometry {
			updateDesktopLayout(proxy)
			change = true
		} else if event == x11.DesktopStacking {
			change = updateWindowList(proxy)
		} else {
			// So it's a 'single'-window event
			if win := getWindow(wId); win != nil {
				var updatedWindow = Window{
					Id:       win.Id,
					self:     win.self,
					Name:     win.Name,
					IconName: win.IconName,
					State:    win.State,
					Stacking: win.Stacking,
				}
				switch event {
				case x11.WindowTitle:
					updatedWindow.Name, _ = x11.GetName(proxy, wId)

				case x11.WindowIconName:
					updatedWindow.IconName, _ = GetIconName(proxy, wId)
				case x11.WindowSt:
					updatedWindow.State = x11.GetStates(proxy, wId)
				// TODO case x11.WindowGeometry:
				default:
					continue
				}
				updateWindow(&updatedWindow)
				change = change || relevantForDesktopSearch(&updatedWindow)
			}
		}

		if change {
			watch.DesktopSearchMayHaveChanged()
		}
	}
}

var (
	repo          sync.Mutex
	windows       []*Window
	desktopLayout *DesktopLayout
)

func getWindow(wId uint32) *Window {
	repo.Lock()
	defer repo.Unlock()
	for _, w := range windows {
		if w.Id == wId {
			return w
		}
	}
	return nil
}

func updateWindow(w *Window) {
	repo.Lock()
	defer repo.Unlock()
	for i, win := range windows {
		if win.Id == w.Id {
			windows[i] = w
			break
		}
	}
}

func getWindows() []*Window {
	repo.Lock()
	defer repo.Unlock()
	var res = make([]*Window, len(windows))
	copy(res, windows)
	return res
}

func getDesktopLayout() *DesktopLayout {
	repo.Lock()
	defer repo.Unlock()

	return desktopLayout
}

func setDesktopLayout(newDesktopLayout *DesktopLayout) {
	repo.Lock()
	defer repo.Unlock()
	desktopLayout = newDesktopLayout
}

func updateWindowList(p x11.Proxy) bool {
	var wIds = x11.GetStack(p)
	var newWindows = make([]*Window, 0, len(wIds))
	var somethingChanged bool
	repo.Lock()
	defer repo.Unlock()

	for i, wId := range wIds {
		var newWindow *Window
		for _, win := range windows {
			if wId == win.Id {
				newWindow = win.shallowCopy()
				break
			}
		}
		if newWindow == nil {
			newWindow = makeWindow(p, wId)
			x11.SubscribeToWindowEvents(p, wId)
		}

		if newWindow.Stacking != i && relevantForDesktopSearch(newWindow) {
			somethingChanged = true
		}

		newWindow.Stacking = i
		newWindows = append(newWindows, newWindow)
	}
	if len(newWindows) < len(windows) {
		somethingChanged = true
	}

	windows = newWindows
	return somethingChanged
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

// --------------------------- Serving http requests -------------------------------

// http requests are concurrent, so all access to x11 from handling an http request, happens through
// this
var requestProxy = x11.MakeProxy()

// - and uses this for synchronization
var requestProxyMutex sync.Mutex

func GetResource(relPath []string) resource.Resource {
	if len(relPath) == 1 {
		if relPath[0] == "desktoplayout" {
			return getDesktopLayout()
		} else if id, err := strconv.ParseUint(relPath[0], 10, 32); err == nil {
			if win := getWindow(uint32(id)); win != nil {
				return win
			}
		}
	}
	return nil
}

func Collect(term string, sink chan resource.Link) {
	for _, win := range getWindows() {
		if relevantForDesktopSearch(win) {
			var rnk = -1
			if term == "" {
				rnk = win.Stacking
			} else {
				rnk = searchutils.Match(term, win.Name)
			}
			if rnk > -1 {
				sink <- resource.MakeRankedLink(win.self, win.Name, win.IconName, "window", rnk)
			}
		}
	}
}

func CollectPaths(method string, sink chan string) {
	for _, win := range getWindows() {
		sink <- win.self
	}
}
