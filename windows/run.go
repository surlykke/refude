// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/respond"
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
		event, wId := x11.NextEvent(proxy)
		var change = false
		switch event {
		case x11.DesktopGeometry:
			updateDesktopLayout(proxy)
			change = true
		case x11.DesktopStacking:
			if repo.getHighlighedWindowId() == 0 {
				change = updateWindowList(proxy)
			}
		case x11.WindowTitle:
			if w, ok := repo.getWindowCopy(wId); ok {
				w.Name, _ = x11.GetName(proxy, wId)
				w.Self.Title = w.Name
				repo.setWindow(w)
				change = relevantForDesktopSearch(w)
			}
		case x11.WindowIconName:
			if w, ok := repo.getWindowCopy(wId); ok {
				w.IconName, _ = GetIconName(proxy, wId)
				w.Self.Icon = icons.IconUrl(w.IconName)
				repo.setWindow(w)
				change = relevantForDesktopSearch(w)
			}
		case x11.WindowSt:
			if w, ok := repo.getWindowCopy(wId); ok {
				var desktopLayout = repo.getDesktopLayout()
				w.State = x11.GetStates(proxy, wId)
				updateLinks(w, desktopLayout)
				if w.State&x11.HIDDEN > 0 {
					w.Self.Traits = []string{"window", "minimized"}
				} else {
					w.Self.Traits = []string{"window"}
				}
				repo.setWindow(w)
				change = relevantForDesktopSearch(w)
			}
		case x11.WindowGeometry:
			// TODO updateWindowGeometry(proxy, wId)
		}

		if change && repo.getHighlighedWindowId() == 0 {
			watch.DesktopSearchMayHaveChanged()
		}
	}
}

type Repo struct {
	sync.Mutex
	windows             map[uint32]*Window
	desktopLayout       *DesktopLayout
	highlightedWindowId uint32
	highlightDeadline   time.Time
}

func (r *Repo) getWindow(wId uint32) (*Window, bool) {
	r.Lock()
	defer r.Unlock()
	w, ok := r.windows[wId]
	return w, ok
}

func (r *Repo) getWindowCopy(wId uint32) (*Window, bool) {
	if w, ok := r.getWindow(wId); !ok {
		return nil, false
	} else {
		return w.copy(), true
	}
}

func (r *Repo) setWindow(w *Window) {
	r.Lock()
	defer r.Unlock()
	r.windows[w.Id] = w
}

func (r *Repo) getWindows() []*Window {
	repo.Lock()
	defer repo.Unlock()
	return extractList(repo.windows)
}

func (r *Repo) getDesktopLayout() *DesktopLayout {
	repo.Lock()
	defer repo.Unlock()

	return repo.desktopLayout
}

func (r *Repo) setDesktopLayout(desktopLayout *DesktopLayout) {
	repo.Lock()
	defer repo.Unlock()
	repo.desktopLayout = desktopLayout
}

func (r *Repo) getHighlighedWindowId() uint32 {
	repo.Lock()
	defer repo.Unlock()
	return repo.highlightedWindowId
}

func (r *Repo) setHighlighedWindowId(wId uint32) {
	repo.Lock()
	defer repo.Unlock()
	repo.highlightedWindowId = wId
}

func (r *Repo) getHighlightDeadline() time.Time {
	repo.Lock()
	defer repo.Unlock()
	return repo.highlightDeadline
}

var repo = &Repo{
	windows:       make(map[uint32]*Window),
	desktopLayout: &DesktopLayout{},
}

func updateWindowList(p x11.Proxy) bool {
	var wIds = x11.GetStack(p)
	var newWindowMap = make(map[uint32]*Window, len(wIds))
	var somethingChanged bool
	repo.Lock()
	defer repo.Unlock()
	for i, wId := range wIds {
		var window *Window
		var ok bool
		if window, ok = repo.windows[wId]; ok {
			window = window.copy()
		} else {
			window = BuildWindow(p, wId)
			updateLinks(window, repo.desktopLayout)
			x11.SubscribeToWindowEvents(p, wId)
		}

		if window.Stacking != i && relevantForDesktopSearch(window) {
			somethingChanged = true
		}

		window.Stacking = i
		newWindowMap[wId] = window
	}

	for _, window := range repo.windows {
		if relevantForDesktopSearch(window) {
			if _, ok := newWindowMap[window.Id]; !ok {
				somethingChanged = true
			}
		}
	}

	repo.windows = newWindowMap

	return somethingChanged
}

func updateDesktopLayout(p x11.Proxy) {
	var monitors = x11.GetMonitorDataList(p)
	var layout = &DesktopLayout{
		Monitors: monitors,
	}
	layout.Resource = respond.MakeResource("/desktoplayout", "DesktopLayout", "", &layout, "desktoplayout")

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

	repo.Lock()
	defer repo.Unlock()
	var newMap = make(map[uint32]*Window, len(repo.windows))
	for id, win := range repo.windows {
		var newWin = win.copy()
		updateLinks(newWin, layout)
		newMap[id] = newWin
	}
	repo.desktopLayout = layout
	repo.windows = newMap
}

// --------------------------- Serving http requests -------------------------------

// http requests are concurrent, so all access to x11 from handling an http request, happens through
// this
var requestProxy = x11.MakeProxy()

// - and is synchronized through this
var requestProxyMutex sync.Mutex

func DesktopLayoutHandler(r *http.Request) http.Handler {
	repo.Lock()
	defer repo.Unlock()
	return repo.desktopLayout
}

var windowPath = regexp.MustCompile("^/window/(\\d+)(/screenshot)?$")

func WindowHandler(r *http.Request) http.Handler {
	if r.URL.Path == "/window/unhighlight" {
		return Unhighligher{}
	} else if matches := windowPath.FindStringSubmatch(r.URL.Path); matches == nil {
		return nil
	} else if val, err := strconv.ParseUint(matches[1], 10, 32); err != nil {
		return nil
	} else {
		var id = uint32(val)
		var screenShot = matches[2] != ""
		if win, ok := repo.getWindow(id); ok {
			if screenShot {
				return ScreenShot(id)
			} else {
				return win
			}
		}
	}
	return nil
}

type Unhighligher struct{}

func (u Unhighligher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		unHighligt()
		respond.Accepted(w)
	} else {
		respond.NotAllowed(w)
	}
}

func DesktopSearch(term string, baserank int) []respond.Link {
	var windowList = repo.getWindows()
	sort.Sort(WindowStack(windowList))
	var links = make([]respond.Link, 0, len(windowList))
	for _, win := range windowList {
		if relevantForDesktopSearch(win) {
			if rank, ok := searchutils.Rank(strings.ToLower(win.Name), term, baserank); ok {
				links = append(links, win.GetRelatedLink(rank))
			}
		}
	}

	return links
}

func AllPaths() []string {
	var windowList = repo.getWindows()
	var paths = make([]string, 0, 2*len(windowList)+3)
	for _, window := range windowList {
		paths = append(paths, fmt.Sprintf("/window/%d", window.Id))
		paths = append(paths, fmt.Sprintf("/window/%d/screenshot", window.Id))
	}
	paths = append(paths, "/windows")
	paths = append(paths, "/monitors")
	paths = append(paths, "/desktoplayout")
	return paths
}

func MakeRaiser(wId uint32) func(*http.Request) error {
	return func(*http.Request) error {
		requestProxyMutex.Lock()
		defer requestProxyMutex.Unlock()
		x11.RaiseAndFocusWindow(requestProxy, wId)
		return nil
	}
}

func MakeRestorer(wId uint32) func(*http.Request) error {
	return func(*http.Request) error {
		requestProxyMutex.Lock()
		defer requestProxyMutex.Unlock()
		x11.RemoveStates(requestProxy, wId, x11.HIDDEN|x11.MAXIMIZED_HORZ|x11.MAXIMIZED_VERT)
		return nil
	}
}

func MakeMinimizer(wId uint32) func(*http.Request) error {
	return func(*http.Request) error {
		requestProxyMutex.Lock()
		defer requestProxyMutex.Unlock()
		x11.AddStates(requestProxy, wId, x11.HIDDEN)
		return nil
	}
}

func MakeMaximizer(wId uint32) func(*http.Request) error {
	return func(*http.Request) error {
		requestProxyMutex.Lock()
		defer requestProxyMutex.Unlock()
		x11.AddStates(requestProxy, wId, x11.MAXIMIZED_HORZ|x11.MAXIMIZED_VERT)
		return nil
	}
}

func MakeMover(wId uint32, monitorName string) func(r *http.Request) error {
	return func(*http.Request) error {
		for _, m := range repo.getDesktopLayout().Monitors {
			if monitorName == m.Name {
				var marginW, marginH = m.W / 10, m.H / 10
				requestProxyMutex.Lock()
				defer requestProxyMutex.Unlock()
				var saveStates = x11.GetStates(requestProxy, wId) & (x11.HIDDEN | x11.MAXIMIZED_HORZ | x11.MAXIMIZED_VERT)
				x11.RemoveStates(requestProxy, wId, x11.HIDDEN|x11.MAXIMIZED_HORZ|x11.MAXIMIZED_VERT)
				x11.SetBounds(requestProxy, wId, m.X+int32(marginW), m.Y+int32(marginH), m.W-2*marginW, m.H-2*marginH)
				x11.AddStates(requestProxy, wId, saveStates)
				return nil
			}
		}
		return fmt.Errorf("Monitor '%s' not found", monitorName)
	}
}

func MakeHighlighter(wId uint32) func(r *http.Request) error {
	return func(*http.Request) error {
		highlighWindow(wId)
		return nil
	}
}

const highlightTimeout = 3 * time.Second
const OPACITY uint32 = 0x11111111

func highlighWindow(wId uint32) {
	repo.Lock()
	requestProxyMutex.Lock()
	defer requestProxyMutex.Unlock()
	defer repo.Unlock()

	if repo.highlightedWindowId == 0 {
		repo.highlightDeadline = time.Now().Add(highlightTimeout)
		scheduleUnhighlight(repo.highlightDeadline)
		for _, win := range repo.windows {
			if win.Id != wId && win.State&(x11.HIDDEN|x11.ABOVE) == 0 {
				x11.SetTransparent(requestProxy, win.Id, OPACITY)
			}
		}

	} else {
		repo.highlightDeadline = time.Now().Add(highlightTimeout)
		x11.SetTransparent(requestProxy, repo.highlightedWindowId, OPACITY)
	}
	x11.SetOpaque(requestProxy, wId)
	x11.RaiseWindow(requestProxy, wId)
	repo.highlightedWindowId = wId
}

func unHighligt() {
	repo.Lock()
	requestProxyMutex.Lock()
	defer requestProxyMutex.Unlock()
	defer repo.Unlock()

	if repo.highlightedWindowId != 0 {
		var list = extractList(repo.windows)
		sort.Sort(WindowStack(list))

		for i := len(list) - 1; i >= 0; i-- {
			x11.SetOpaque(requestProxy, list[i].Id)
			x11.RaiseWindow(requestProxy, list[i].Id)
		}
		repo.highlightedWindowId = 0
	}
}

func scheduleUnhighlight(at time.Time) {
	time.AfterFunc(at.Sub(time.Now())+100*time.Millisecond, func() {
		if repo.getHighlightDeadline().After(time.Now()) {
			scheduleUnhighlight(repo.highlightDeadline)
		} else {
			unHighligt()
		}
	})
}

func updateWindowGeometry(p x11.Proxy, wId uint32) {
	// TODO
}

func extractList(wm map[uint32]*Window) []*Window {
	var list = make([]*Window, 0, len(wm))
	for _, w := range wm {
		list = append(list, w)
	}
	return list
}
