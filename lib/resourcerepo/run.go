package resourcerepo

import (
	"net/http"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"golang.org/x/exp/slices"
)

var lock sync.Mutex
var repo = make(map[string]resource.Resource)

func Put(res resource.Resource) {
	lock.Lock()
	defer lock.Unlock()
	repo[res.Base().Path] = res
}

func Update(res resource.Resource) {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := repo[res.Base().Path]; ok {
		repo[res.Base().Path] = res
	}
}


func Get(path string) (resource.Resource, bool) {
	lock.Lock()
	defer lock.Unlock()
	res, ok := repo[path]
	return res, ok
}

func GetAll() []resource.Resource {
	lock.Lock()
	defer lock.Unlock()
	var all = make([]resource.Resource, 0, len(repo))
	for _, res := range repo {
		all = append(all, res)
	}
	return all
}


func GetTyped[T resource.Resource](path string) (T, bool) {
	// Calls Get, so no lock
	if res, ok := Get(path); ok {
		if t, ok := res.(T); ok {
			return t, true
		}
 	}
	var t T
	return t, false
 }


func GetByPrefix(prefix string) []resource.Resource {
	lock.Lock()
	defer lock.Unlock()
	var result = make([]resource.Resource, 0, 50)
	for path, res := range repo {
		if strings.HasPrefix(path, prefix) {
			result = append(result, res)
		}
	}
	return result
}

func GetTypedByPrefix[T resource.Resource](prefix string) []T {
	var list = make([]T, 0, 20)
	lock.Lock()
	defer lock.Unlock()
	for path, res := range repo {
		if strings.HasPrefix(path, prefix) {
			if t, ok := res.(T); ok {
				list = append(list, t)
			}
		}
	}
	return list
}

func GetTypedAndSortedByPrefix[T resource.Resource](prefix string, reverse bool) []T {
	var list = GetTypedByPrefix[T](prefix)
	if reverse {
		slices.SortFunc(list, func(t1, t2 T) bool { return strings.Compare(t1.Base().Path, t2.Base().Path) > 0})
	} else { 
		slices.SortFunc(list, func(t1, t2 T) bool { return strings.Compare(t1.Base().Path, t2.Base().Path) < 0})
	}
	return list
}


func FindTypedUnderPrefix[T resource.Resource](prefix string, test func(t T) bool) []T {
	lock.Lock()
	defer lock.Unlock()
	var result = []T{}
	for path, res := range repo {
		if strings.HasPrefix(path, prefix) {
			if t, ok := res.(T); ok && test(t) {
				result = append(result, t)
			}
		}
	}
	return result
}

func ReplacePrefixWithList[T resource.Resource](prefix string, newResources []T) {
	lock.Lock()
	defer lock.Unlock()
	for path := range repo {
		if strings.HasPrefix(path, prefix) {
			delete(repo, path)
		}
	}
	for _, res := range newResources {
		repo[res.Base().Path] = res
	}
}

/*
 * Removes all entries having prefix as prefix of key.
 * And then adds all members of map 
 */
func ReplacePrefixWithMap[T resource.Resource](prefix string, newResources map[string]T) {
	lock.Lock()
	defer lock.Unlock()
	for path := range repo {
		if strings.HasPrefix(path, prefix) {
			delete(repo, path)
		}
	}
	for _, res := range newResources {
		repo[res.Base().Path] = res
	}
}

func Remove(path string) {
	lock.Lock()
	defer lock.Unlock()
	delete(repo, path)
}		

func Search(term string) []resource.Resource {
	var all = GetAll()
	var links = make([]resource.Resource, 0, len(all))
	for _, res :=  range all {
		if res.RelevantForSearch(term) && searchutils.Match(term, res.Base().Title) >= 0 {
			links = append(links, res)
		}
	}
	// FIXME sort by rank...
	return links
}

func GetPaths() []string {
	lock.Lock()
	defer lock.Unlock()
	var paths = make([]string, 0, len(repo))
	for path := range repo {
		paths = append(paths, path)
	}
	return paths
}


func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path 
	if strings.HasSuffix(path, "/") {
		resource.ServeList(w, r, GetByPrefix(path))
	} else if res, ok := Get(path); !ok {
		respond.NotFound(w)
	} else {
		resource.ServeSingleResource(w, r, res)
	}
}

