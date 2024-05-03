package resourcerepo

import (
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/stringhash"
)

var lock sync.Mutex
var repo = make(map[string]resource.Resource)

func Put(res resource.Resource) {
	lock.Lock()
	defer lock.Unlock()
	repo[res.Data().Path] = res
}

func Update(res resource.Resource) {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := repo[res.Data().Path]; ok {
		repo[res.Data().Path] = res
	}
}

func Get(path string) (resource.Resource, bool) {
	lock.Lock()
	defer lock.Unlock()
	res, ok := repo[path]
	return res, ok
}

func GetTyped[T resource.Resource](path string) (T, bool) {
	// Calls Get, so no lock
	if res, ok := Get(path); ok {
		return res.(T), true
	}
	var t T
	return t, false
}

func GetByPrefixes(prefixes ...string) []resource.Resource {
	lock.Lock()
	defer lock.Unlock()
	var result = make([]resource.Resource, 0, 50)
	for path, res := range repo {
		for _, prefix := range prefixes {
			if strings.HasPrefix(path, prefix) {
				result = append(result, res)
				break
			}
		}
	}
	slices.SortFunc(result, func(r1, r2 resource.Resource) int { return strings.Compare(r1.Data().Path, r2.Data().Path)})
	return result

}

func GetTypedByPrefix[T resource.Resource](prefix string) []T {
	var resources = GetByPrefixes(prefix) 
	var typed = make([]T, 0, len(resources))
	for _, res := range resources {
		typed = append(typed, res.(T))
	}
	return typed	
}


/*
 * Removes all entries having prefix as prefix of key.
 * And then adds all members of list 
 */
func ReplacePrefixWithList[T resource.Resource](prefix string, newResources []T) {
	lock.Lock()
	defer lock.Unlock()
	for path := range repo {
		if strings.HasPrefix(path, prefix) {
			delete(repo, path)
		}
	}
	for _, res := range newResources {
		repo[res.Data().Path] = res
	}
}

/*
 * As above
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
		repo[res.Data().Path] = res
	}
}

func Remove(path string) {
	lock.Lock()
	defer lock.Unlock()
	delete(repo, path)
}

func RepoHash() uint64 {
	var hash uint64 = 0
	lock.Lock()
	defer lock.Unlock()

	for _, res := range repo {
		if !res.Data().HideFromSearch {
			hash = hash ^ stringhash.FNV1a(res.Data().Title, res.Data().IconUrl)
		}
	}
	return hash
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
		resource.ServeList(w, r, GetByPrefixes(path))
	} else if res, ok := Get(path); !ok {
		respond.NotFound(w)
	} else {
		resource.ServeSingleResource(w, r, res)
	}
}
