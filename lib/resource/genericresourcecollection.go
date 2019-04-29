package resource

import (
	"sort"
	"strings"
	"sync"
)

type GenericResourceCollection struct {
	sync.Mutex
	collectionResources map[string][]string // Not thread safe, must be set before serving begins
	resources           map[string]Resource
}

type ResourceCond func(resource Resource) bool

func MakeGenericResourceCollection() *GenericResourceCollection {
	var grc = &GenericResourceCollection{}
	grc.collectionResources = map[string][]string{}
	grc.resources = make(map[string]Resource)
	return grc
}

// When embedding
func (gcr *GenericResourceCollection) InitializeGenericResourceCollection() {
	gcr.collectionResources = make(map[string][]string)
	gcr.resources = make(map[string]Resource)
}

func (grc *GenericResourceCollection) AddCollectionResource(path string, prefixes ...string) *GenericResourceCollection {
	grc.collectionResources[path] = append(grc.collectionResources[path], prefixes...)
	return grc
}

func (grc *GenericResourceCollection) Get(path string) Resource {
	grc.Lock()
	defer grc.Unlock()

	return grc.resources[path]
}

func (grc *GenericResourceCollection) GetList(path string) []Resource {
	if prefixes, ok := grc.collectionResources[path]; ok {
		var resList = make([]Resource, 0, len(grc.resources))
		grc.Lock()
		defer grc.Unlock()

		for _, prefix := range prefixes {
			for path, resource := range grc.resources {
				if strings.HasPrefix(path, prefix) {
					resList = append(resList, resource)
				}
			}
		}

		sort.Sort(ResourceList(resList))

		return resList
	} else {
		return nil
	}

}

func (grc *GenericResourceCollection) Set(resource Resource) {
	grc.Lock()
	defer grc.Unlock()

	grc.resources[string(resource.GetSelf())] = resource
}

func (grc *GenericResourceCollection) ReplaceAll(newcollection map[string]Resource) {
	grc.Lock()
	defer grc.Unlock()

	grc.resources = newcollection
}

func (grc *GenericResourceCollection) Remove(path string) bool {
	grc.Lock()
	defer grc.Unlock()

	if _, ok := grc.resources[path]; ok {
		delete(grc.resources, path)
		return true
	}
	return false
}

func (grc *GenericResourceCollection) RemoveIf(path string, cond ResourceCond) bool {
	grc.Lock()
	defer grc.Unlock()

	if res, ok := grc.resources[path]; ok && cond(res) {
		delete(grc.resources, path)
		return true
	}
	return false
}

func (gcr *GenericResourceCollection) Filter(cond func(Resource) bool) []Resource {
	var resList = make([]Resource, 0, len(gcr.resources))
	gcr.Lock()
	defer gcr.Unlock()
	for _, res := range gcr.resources {
		if cond(res) {
			resList = append(resList, res)
		}
	}
	return resList
}

func isProperPrefix(prefix, s string) bool {
	return len(prefix) < len(s) && s[0:len(prefix)] == prefix
}
