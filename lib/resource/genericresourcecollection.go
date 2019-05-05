package resource

import (
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/requests"
)

type resourceMap map[string]Resource

type GenericResourceCollection struct {
	sync.Mutex
	resourceMap
}

type ResourceCond func(Resource) bool

type PathCond func(string) bool

func MakeGenericResourceCollection() *GenericResourceCollection {
	return &GenericResourceCollection{resourceMap: make(map[string]Resource)}
}

func (grc *GenericResourceCollection) Get(path string) Resource {
	grc.Lock()
	defer grc.Unlock()

	return grc.resourceMap[path]
}

func (grc *GenericResourceCollection) Set(path string, resource Resource) {
	grc.Lock()
	defer grc.Unlock()

	grc.resourceMap[path] = resource
}

func (grc *GenericResourceCollection) GetByPrefix(prefix string) []Resource {
	grc.Lock()
	defer grc.Unlock()

	var list = make([]Resource, 0, len(grc.resourceMap))
	for path, res := range grc.resourceMap {
		if strings.HasPrefix(path, prefix) {
			list = append(list, res)
		}
	}

	return list
}

type Entry struct {
	path string
	res  Resource
}

func (grc *GenericResourceCollection) GetByCond(cond PathCond) []Entry {
	grc.Lock()
	defer grc.Unlock()

	var list = make([]Entry, 0, len(grc.resourceMap))
	for path, res := range grc.resourceMap {
		if cond(path) {
			list = append(list, Entry{path, res})
		}
	}

	return list
}

func (grc *GenericResourceCollection) ReplaceAll(newcollection map[string]Resource) {
	grc.Lock()
	defer grc.Unlock()

	grc.resourceMap = newcollection
}

func (grc *GenericResourceCollection) Remove(path string) bool {
	grc.Lock()
	defer grc.Unlock()

	if _, ok := grc.resourceMap[path]; ok {
		delete(grc.resourceMap, path)
		return true
	}
	return false
}

func (grc *GenericResourceCollection) RemoveIf(path string, cond ResourceCond) bool {
	grc.Lock()
	defer grc.Unlock()

	if res, ok := grc.resourceMap[path]; ok && cond(res) {
		delete(grc.resourceMap, path)
		return true
	}
	return false
}

type Collection struct {
	repo *GenericResourceCollection
	cond PathCond
}

func (grc *GenericResourceCollection) MakePrefixCollection(prefix string) *Collection {
	return &Collection{
		repo: grc,
		cond: func(path string) bool {
			return strings.HasPrefix(path, prefix)
		}}
}

type SortableEntryList []Entry

func (sel SortableEntryList) Len() int           { return len(sel) }
func (sel SortableEntryList) Swap(i, j int)      { sel[i], sel[j] = sel[j], sel[i] }
func (sel SortableEntryList) Less(i, j int) bool { return sel[i].path < sel[j].path }

func (grc *GenericResourceCollection) MakeRegexpCollection(reg *regexp.Regexp) *Collection {
	return &Collection{
		repo: grc,
		cond: func(path string) bool {
			return reg.MatchString(path)
		}}
}

func (rc Collection) ServeHttp(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var entries = rc.repo.GetByCond(rc.cond)

	var matcher, err = requests.GetMatcher(r)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if matcher != nil {
		var matched = 0
		for i := 0; i < len(entries); i++ {
			if matcher(entries[i].res) {
				entries[matched] = entries[i]
				matched++
			}
		}
		entries = entries[0:matched]
	}

	sort.Sort(SortableEntryList(entries))

	var bytes []byte
	var etag string

	_, brief := r.URL.Query()["brief"]
	if brief {
		var briefs = make([]string, len(entries), len(entries))
		for i := 0; i < len(entries); i++ {
			briefs[i] = entries[i].path
		}
		bytes, etag = ToBytesAndEtag(briefs)
	} else {
		var resources = make([]Resource, len(entries), len(entries))
		for i, entry := range entries {
			resources[i] = entry.res
		}
		bytes, etag = ToBytesAndEtag(resources)
	}

	if statusCode := requests.CheckEtag(r, etag); statusCode != 0 {
		w.WriteHeader(statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", etag)
	_, _ = w.Write(bytes)

}
