package power

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

func Handler(r *http.Request) http.Handler {
	if device := getDevice(r.URL.Path); device != nil {
		return device
	} else {
		return nil
	}
}

func DesktopSearch(term string, baserank int) []respond.Link {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	var links = make([]respond.Link, 0, len(devices))
	for _, device := range devices {
		if deviceSelf(device.DbusPath) != "/device/DisplayDevice" && device.Type != "Line Power" {
			if rank, ok := searchutils.Rank(device.Self.Title, term, baserank); ok {
				links = append(links, device.GetRelatedLink(rank))
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
