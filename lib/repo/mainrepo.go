package repo

import (
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
)

var resources = make(map[string]resource.Resource, 200)
var lock sync.Mutex

func Put(res resource.Resource) {
	lock.Lock()
	defer lock.Unlock()
	resources[res.GetPath()] = res
}

func Remove(path string) {
	lock.Lock()
	defer lock.Unlock()
	delete(resources, path)
}

func Replace(resList []resource.Resource, prefix string) {
	lock.Lock()
	defer lock.Unlock()
	for path := range resources {
		if strings.HasPrefix(path, prefix) {
			delete(resources, path)
		}
	}
	for _, res := range resList {
		resources[res.GetPath()] = res
	}
}

func GetUntyped(path string) resource.Resource {
	lock.Lock()
	defer lock.Unlock()
	return resources[path]
}

func Get[T resource.Resource](path string) (T, bool) {
	if res := GetUntyped(path); res != nil {
		if t, ok := res.(T); ok {
			return t, true
		}
	}
	var t T
	return t, false
}

func GetListUntyped(prefix string) []resource.Resource {
	lock.Lock()
	defer lock.Unlock()
	var resList = make([]resource.Resource, 0, 100)
	for _, res := range resources {
		if strings.HasPrefix(res.GetPath(), prefix) {
			resList = append(resList, res)
		}
	}
	return resList
}

func GetList[T resource.Resource](prefix string) []T {
	var resList = GetListUntyped(prefix)
	var list = make([]T, 0, len(resList))
	for _, res := range resList {
		if t, ok := res.(T); ok {
			list = append(list, t)
		}
	}
	return list
}

func GetListSortedByPath[T resource.Resource](prefix string) []T {
	var resList = GetList[T](prefix)
	slices.SortFunc(resList, func(t1, t2 T) int { return strings.Compare(t1.GetPath(), t2.GetPath()) })
	return resList
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	if strings.HasSuffix(path, "/") {
		resource.ServeList(w, r, GetListUntyped(path))
	} else {
		if res := GetUntyped(path); res != nil {
			resource.ServeSingleResource(w, r, res)
		} else {
			respond.NotFound(w)
		}
	}
}
