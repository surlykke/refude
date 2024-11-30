package repo

import (
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/path"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
)

var resources = make(map[path.Path]resource.Resource, 200)
var lock sync.Mutex

func Put(res resource.Resource) {
	lock.Lock()
	defer lock.Unlock()
	resources[res.Data().Path] = res
}

func Remove(path path.Path) {
	lock.Lock()
	defer lock.Unlock()
	delete(resources, path)
}

func RemoveTyped[T resource.Resource](path path.Path) (T, bool) {
	lock.Lock()
	defer lock.Unlock()
	if r, ok := resources[path]; ok {
		if t, ok := r.(T); ok {
			delete(resources, path)
			return t, true
		}
	}
	var t T
	return t, false
}

func Replace(resList []resource.Resource, prefix string) {
	lock.Lock()
	defer lock.Unlock()
	for path := range resources {
		if strings.HasPrefix(string(path), prefix) {
			delete(resources, path)
		}
	}
	for _, res := range resList {
		resources[res.Data().Path] = res
	}
}

func GetUntyped(path path.Path) resource.Resource {
	lock.Lock()
	defer lock.Unlock()
	return resources[path]
}

func Get[T resource.Resource](path path.Path) (T, bool) {
	if res := GetUntyped(path); res != nil {
		if t, ok := res.(T); ok {
			return t, true
		}
	}
	var t T
	return t, false
}

func GetListUntyped(prefixes ...string) []resource.Resource {
	lock.Lock()
	defer lock.Unlock()
	var resList = make([]resource.Resource, 0, 100)
	for _, prefix := range prefixes {
		for _, res := range resources {
			if strings.HasPrefix(string(res.Data().Path), prefix) {
				resList = append(resList, res)
			}
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

func CollectLinks(prefix string) []resource.Link {
	var resList = GetListUntyped(prefix)
	var result = make([]resource.Link, 0, len(resList))
	for _, res := range resList {
		if !res.OmitFromSearch() {
			result = append(result, res.Data().Link())
		}
	}
	return result
}

func GetListSortedByPath[T resource.Resource](prefix string) []T {
	var resList = GetList[T](prefix)
	slices.SortFunc(resList, func(t1, t2 T) int { return strings.Compare(string(t1.Data().Path), string(t2.Data().Path)) })
	return resList
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var reqPath = r.URL.Path
	if strings.HasSuffix(string(reqPath), "/") {
		resource.ServeList(w, r, GetListUntyped(reqPath))
	} else {
		if res := GetUntyped(path.Of(reqPath)); res != nil {
			resource.ServeSingleResource(w, r, res)
		} else {
			respond.NotFound(w)
		}
	}
}
