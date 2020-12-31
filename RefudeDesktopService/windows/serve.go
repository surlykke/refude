package windows

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/lib/slice"
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
		for _, wId := range windows.Load().([]uint32) {
			if id == wId {
				if screenShot {
					return ScreenShot(id)
				} else {
					return Window(id)
				}
			}
		}

	}
	return nil
}

func Windows() respond.Links {
	var idList = windows.Load().([]uint32)
	var links = make(respond.Links, 0, len(idList))
	for _, id := range idList {
		links = append(links, Window(id).ToData().Link())
	}
	return links
}

func DesktopSearch(term string, baserank int) respond.Links {
	var idList = windows.Load().([]uint32)
	var links = make(respond.Links, 0, len(idList))
	for _, id := range idList {
		var wd = Window(id).ToData()
		if slice.Contains(wd.States, "_NET_WM_STATE_ABOVE", "_NET_WM_STATE_SKIP_TASKBAR") {
			continue
		}

		if rank, ok := searchutils.Rank(strings.ToLower(wd.Name), term, baserank); ok {
			var link = wd.Link()
			link.Rank = rank
			links = append(links, link)
		}
	}

	return links
}

func AllPaths() []string {
	var windowList = windows.Load().([]uint32)
	var monitorList = monitors.Load().([]*Monitor)
	var paths = make([]string, 0, 2*len(windowList)+len(monitorList)+2)
	for _, window := range windowList {
		paths = append(paths, fmt.Sprintf("/window/%d", window))
		paths = append(paths, fmt.Sprintf("/window/%d/screenshot", window))
	}
	for _, monitor := range monitorList {
		paths = append(paths, monitor.Link().Href)
	}
	paths = append(paths, "/windows")
	paths = append(paths, "/monitors")
	return paths
}
