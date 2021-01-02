package power

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

func Handler(r *http.Request) http.Handler {
	if r.URL.Path == "/devices" {
		return Collect()
	} else if device := getDevice(r.URL.Path); device != nil {
		return device
	} else {
		return nil
	}
}

func DesktopSearch(term string, baserank int) respond.Links {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	var links = make(respond.Links, 0, len(devices))
	for _, device := range devices {
		if deviceSelf(device.DbusPath) != "/device/DisplayDevice" && device.Type != "Line Power" {
			var link = device.Link()
			var ok bool
			if link.Rank, ok = searchutils.Rank(link.Title, term, baserank); ok {
				links = append(links, link)
			}
		}
	}

	return links
}

func AllPaths() []string {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	var paths = make([]string, 0, len(devices)+1)
	for path := range devices {
		paths = append(paths, path)
	}
	paths = append(paths, "/devices")
	return paths
}
