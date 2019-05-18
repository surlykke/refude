package resource

import (
	"net/http"
	"sync"

	"github.com/surlykke/RefudeServices/lib/requests"
)

// A resource is something that can handle an incoming http request, and has an etag
type Res interface {
	http.Handler
	GetEtag() string // may be empty
}

type ResourceMap struct {
	resources map[string]Res
	sync.Mutex
	sync.Cond
}

func MakeResourceMap() *ResourceMap {
	var resMap = &ResourceMap{}
	resMap.resources = make(map[string]Res)
	resMap.L = &sync.Mutex{}
	return resMap
}

func (rm *ResourceMap) Get(path string) Res {
	rm.Lock()
	defer rm.Unlock()
	return rm.resources[path]
}

func (rm *ResourceMap) LongGet(path string, etagList string) Res {
	rm.L.Lock()
	defer rm.L.Unlock()
	for {
		var res = rm.Get(path)
		if res == nil || res.GetEtag() == "" || !requests.EtagMatch(res.GetEtag(), etagList) {
			return res
		}
		rm.Wait()
	}
}

func (rm *ResourceMap) Set(path string, res Res) {
	rm.Lock()
	defer rm.Unlock()

	rm.resources[path] = res
}

func (rm *ResourceMap) Remove(path string) bool {
	rm.Lock()
	defer rm.Unlock()

	if _, ok := rm.resources[path]; ok {
		delete(rm.resources, path)
		return true
	}
	return false
}

func (rm *ResourceMap) ReplaceAll(newcollection map[string]Res) {
	rm.Lock()
	defer rm.Unlock()

	rm.resources = newcollection
}
