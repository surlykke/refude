package power

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/respond"
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

func DesktopSearch() respond.StandardFormatList {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	var res = make(respond.StandardFormatList, 0, len(devices))
	for _, device := range devices {
		if device.self != "/device/DisplayDevice" && device.Type != "Line Power" {
			res = append(res, device.ToStandardFormat())
			continue
		}
	}

	return res
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
