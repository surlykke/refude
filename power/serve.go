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

func Collect(term string, sink chan resource.Link) {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	for _, d := range devices {
		if !d.DisplayDevice {
			if rnk := searchutils.Match(term, d.title); rnk > -1 {
				sink <- resource.MakeRankedLink(d.self, d.title, d.IconName, "device", rnk)
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
