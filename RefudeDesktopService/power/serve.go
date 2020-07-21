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
	for path, device := range devices {
		if path != "/device/DisplayDevice" {
			res = append(res, device.ToStandardFormat())
			continue
		}
	}

	return res
}
