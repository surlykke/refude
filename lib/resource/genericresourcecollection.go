package resource

import (
	"strings"
	"sync"
)

type GenericResourceCollection struct {
	sync.Mutex
	prefixes  []string // maps list paths to prefixes, eg. /windows -> /window/
	resources map[string]Resource
}

type ResourceCond func(resource Resource) bool

func MakeGenericResourceCollection(prefixes ...string) *GenericResourceCollection {
	var grc = &GenericResourceCollection{}
	grc.prefixes = prefixes
	grc.resources = make(map[string]Resource)
	return grc
}

func (grc *GenericResourceCollection) OwnsPath(path string) bool {
	for _, prefix := range grc.prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func (grc *GenericResourceCollection) Get(path string) Resource {
	grc.Lock()
	defer grc.Unlock()

	return grc.resources[path]
}

func (grc *GenericResourceCollection) GetByPrefix(prefix string) []Resource {
	var isAPrefix = false
	for _, p := range grc.prefixes {
		if prefix == p {
			isAPrefix = true
			break
		}
	}

	if !isAPrefix {
		return nil
	}

	grc.Lock()
	defer grc.Unlock()

	var list = make([]Resource, 0, len(grc.resources))
	for path, res := range grc.resources {
		if isProperPrefix(prefix, path) {
			list = append(list, res)
		}
	}

	return list
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

func isProperPrefix(prefix, s string) bool {
	return len(prefix) < len(s) && s[0:len(prefix)] == prefix
}
