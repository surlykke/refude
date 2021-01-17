package windows

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

func MonitorHandler(r *http.Request) http.Handler {
	if r.URL.Path == "/monitors" {
		return monitorLinks()
	} else if strings.HasPrefix(r.URL.Path, "/monitor/") {
		var monitorName = r.URL.Path[len("/monitor/"):]
		for _, m := range monitors.Load().([]*Monitor) {
			if m.Name == monitorName {
				return m
			}
		}
	}
	return nil
}

func monitorLinks() respond.Links {
	var mList = monitors.Load().([]*Monitor)
	var links = make(respond.Links, len(mList), len(mList))
	for i := 0; i < len(mList); i++ {
		links[i] = mList[i].Link()
	}
	return links
}

var windowPath = regexp.MustCompile("^/window/(\\d+)(/screenshot)?$")

func WindowHandler(r *http.Request) http.Handler {
	if r.URL.Path == "/windows" {
		return Windows()
	} else if matches := windowPath.FindStringSubmatch(r.URL.Path); matches == nil {
		return nil
	} else if val, err := strconv.ParseUint(matches[1], 10, 32); err != nil {
		return nil
	} else {
		var id = uint32(val)
		var screenShot = matches[2] != ""
		for _, win := range windows.Load().([]*Window) {
			if id == win.Id {
				if screenShot {
					return ScreenShot(id)
				} else {
					return win
				}
			}
		}

	}
	return nil
}

func Windows() respond.Links {
	var windowList = windows.Load().([]*Window)
	var links = make(respond.Links, len(windowList), len(windowList))
	for i, win := range windowList {
		links[i] = win.Link()
	}
	return links
}

func DesktopSearch(term string, baserank int) respond.Links {
	var windowList = windows.Load().([]*Window)
	var links = make(respond.Links, 0, len(windowList))
	for _, win := range windowList {
		if win.State&(ABOVE|SKIP_TASKBAR) != 0 {
			continue
		}

		if rank, ok := searchutils.Rank(strings.ToLower(win.Name), term, baserank); ok {
			var link = win.Link()
			link.Rank = rank
			links = append(links, link)
		}
	}

	return links
}

func AllPaths() []string {
	var windowList = windows.Load().([]*Window)
	var monitorList = monitors.Load().([]*Monitor)
	var paths = make([]string, 0, 2*len(windowList)+len(monitorList)+2)
	for _, window := range windowList {
		paths = append(paths, fmt.Sprintf("/window/%d", window.Id))
		paths = append(paths, fmt.Sprintf("/window/%d/screenshot", window.Id))
	}
	for _, monitor := range monitorList {
		paths = append(paths, monitor.Link().Href)
	}
	paths = append(paths, "/windows")
	paths = append(paths, "/monitors")
	return paths
}
