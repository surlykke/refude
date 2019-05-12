package resource

import (
	"sort"
	"strings"
	"sync"
)

type Selfie interface {
	GetSelf() string
}

type ResourceMap struct {
	collections map[string]bool
	resources   map[string]interface{}
	sync.Mutex
}

func MakeResourceMap(collections ...string) *ResourceMap {
	var rm = &ResourceMap{
		collections: make(map[string]bool),
		resources:   make(map[string]interface{}),
	}
	for _, collection := range collections {
		rm.AddCollection(collection)
	}
	return rm
}

type resList []interface{}

func (rl resList) Len() int      { return len(rl) }
func (rl resList) Swap(i, j int) { rl[i], rl[j] = rl[j], rl[i] }
func (rl resList) Less(i, j int) bool {
	var sl1, ok1 = rl[i].(Selfie)
	var sl2, ok2 = rl[j].(Selfie)
	return ok1 && ok2 && (sl1.GetSelf() < sl2.GetSelf())
}

func (rm *ResourceMap) Get(path string) interface{} {
	rm.Lock()
	defer rm.Unlock()

	if rm.collections[path] {
		var prefix = path[0:len(path)-1] + "/"
		var prefixLen = len(prefix)

		var list = make([]interface{}, 0, len(rm.resources))
		for p, res := range rm.resources {
			if strings.HasPrefix(p, prefix) && strings.Index(p[prefixLen:], "/") == -1 {
				if selfie, ok := res.(Selfie); ok {
					list = append(list, selfie)
				}
			}
		}

		sort.Sort(resList(list))

		return list
	} else {
		return rm.resources[path]
	}
}

func (rm *ResourceMap) Set(path string, res interface{}) {
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

func (rm *ResourceMap) RemoveIf(path string, cond func(interface{}) bool) bool {
	rm.Lock()
	defer rm.Unlock()

	if res, ok := rm.resources[path]; ok && cond(res) {
		delete(rm.resources, path)
		return true
	}
	return false
}

func (rm *ResourceMap) AddCollection(path string) {
	rm.Lock()
	defer rm.Unlock()

	rm.collections[path] = true
}

func (rm *ResourceMap) ReplaceAll(newcollection map[string]interface{}) {
	rm.Lock()
	defer rm.Unlock()

	rm.resources = newcollection
}
