package resourcerepo

import (
	"net/http"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

var lock sync.Mutex
var repo = make(map[string]resource.Resource)

func Put(res resource.Resource) {
	lock.Lock()
	defer lock.Unlock()
	repo[res.GetPath()] = res
}

func Update(res resource.Resource) {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := repo[res.GetPath()]; ok {
		repo[res.GetPath()] = res
	}
}


func Get(path string) (resource.Resource, bool) {
	lock.Lock()
	defer lock.Unlock()
	res, ok := repo[path]
	return res, ok
}

func GetTyped[T resource.Resource](path string) (T, bool) {
	if res, ok := Get(path); ok {
		if t, ok := res.(T); ok {
			return t, true
		}
 	}
	var t T
	return t, false
 }


func GetByPrefix(prefix string) resource.List {
	lock.Lock()
	defer lock.Unlock()
	var result = make(resource.List, 0, 50)
	for path, res := range repo {
		if strings.HasPrefix(path, prefix) {
			result = append(result, res)
		}
	}
	return result
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
		repo[res.GetPath()] = res
	}
}

func ReplacePrefixWithMap[T resource.Resource](prefix string, newResources map[string]T) {
	lock.Lock()
	defer lock.Unlock()
	for path := range repo {
		if strings.HasPrefix(path, prefix) {
			delete(repo, path)
		}
	}
	for _, res := range newResources {
		repo[res.GetPath()] = res
	}
}

func Remove(path string) {
	lock.Lock()
	defer lock.Unlock()
	delete(repo, path)
}		

func Search(term string) link.List {
	lock.Lock()
	defer lock.Unlock()
	var links = make(link.List, 0, len(repo))
	for _, res := range repo {
		if res.RelevantForSearch(term) {
			var title = res.GetTitle()
			var icon = res.GetIconUrl()
			var profile = res.GetProfile()
			if rnk := searchutils.Match(term, title, res.GetKeywords()...); rnk > -1 {
				links = append(links, link.MakeRanked(res.GetPath(), title, icon, profile, rnk))
			}
		}
	}
	links.SortByRank()
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

