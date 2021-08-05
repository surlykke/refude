package power

import (
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

func GetResource(pathElements []string) resource.Resource {
	if len(pathElements) == 1 {
		if d := getDevice("/device/" + pathElements[0]); d != nil {
			return d
		}
	}
	return nil
}

func Crawl(term string, forDisplay bool, crawler searchutils.Crawler) {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	for _, d := range devices {
		if !d.DisplayDevice {
			crawler(d.self, d.title, d.IconName)
		}
	}
}
