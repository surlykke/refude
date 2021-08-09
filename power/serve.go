package power

import (
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

func GetResource(relPath string) resource.Resource {
	if d := getDevice("/device/" + relPath); d != nil {
		return d
	}
	return nil
}

func Collect(term string, sink chan link.Link) {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	for _, d := range devices {
		if !d.DisplayDevice {
			if rnk := searchutils.Match(term, d.title); rnk > -1 {
				sink <- link.MakeRanked(d.self, d.title, d.IconName, "device", rnk)
			}
		}
	}
}

func CollectPaths(method string, sink chan string) {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	for _, d := range devices {
		sink <- d.self
	}
}
