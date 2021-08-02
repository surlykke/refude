package power

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

func GetJsonResource(r *http.Request) respond.JsonResource {
	if device := getDevice(r.URL.Path); device != nil {
		return device
	} else {
		return nil
	}
}

func Crawl(term string, forDisplay bool, crawler searchutils.Crawler) {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	for _, device := range devices {
		crawler(device.GetRelatedLink(), nil)
	}
}
