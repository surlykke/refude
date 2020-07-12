package power

import (
	"fmt"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/devices" {
		respond.AsJson(w, r, Collect(searchutils.Term(r)))
	} else if device := getDevice(r.URL.Path); device != nil {
		respond.AsJson(w, r, device.ToStandardFormat())
	} else {
		respond.NotFound(w)
	}
}

func DesktopSearch(term string) respond.StandardFormatList {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	var res = make(respond.StandardFormatList, 0, len(devices))
	for path, device := range devices {
		if path == "/device/DisplayDevice" {
			continue
		}
		var sf = device.ToStandardFormat()
		fmt.Printf("Matching '%s' against '%s' and '%s'\n", term, sf.Title, sf.Comment)
		if rank := searchutils.SimpleRank(sf.Title, sf.Comment, term); rank > -1 {
			res = append(res, sf)
		}
	}

	return res
}
