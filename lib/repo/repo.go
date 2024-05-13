package repo

import (
	"fmt"
	"slices"
	"strings"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

// Not threadsafe
type Repo[T resource.Resource] struct {
	resources    map[string]T
	searchFilter func(string, T) bool
}

func MakeRepo[T resource.Resource]() Repo[T] {
	return MakeRepoWithFilter[T](nil)
}

func MakeRepoWithFilter[T resource.Resource](filter func(string, T) bool) Repo[T] {
	return Repo[T]{
		resources:    make(map[string]T),
		searchFilter: filter,
	}
}

func (m *Repo[T]) Put(t T) {
	m.resources[t.Data().Path] = t
}

func (m *Repo[T]) Get(path string) (T, bool) {
	t, ok := m.resources[path]
	return t, ok
}

func (m *Repo[T]) GetAll() []T {
	var all = make([]T, 0, len(m.resources))
	for _, t := range m.resources {
		all = append(all, t)
	}
	slices.SortFunc(all, func(t1, t2 T) int { return strings.Compare(t1.Data().Path, t2.Data().Path) })
	return all
}

func (m *Repo[T]) Remove(path string) {
	delete(m.resources, path)
}

func (m *Repo[T]) RemoveAll() {
	for path := range m.resources {
		delete(m.resources, path)
	}
}

func (m Repo[T]) GetResourcesByPrefix(prefix string) []resource.Resource {
	var resList = make([]resource.Resource, 0, len(m.resources))
	for _, res := range m.resources {
		if strings.HasPrefix(res.Data().Path, prefix) {
			resList = append(resList, res)
		}
	}
	return resList
}

func (m Repo[T]) DoRequest(req ResourceRequest) {
	switch req.ReqType {
	case Search:
		if m.searchFilter != nil {
			for _, res := range m.resources {
				if m.searchFilter(req.Data, res) {
					if rnk := searchutils.Match(req.Data, res.Data().Title, res.Data().Keywords...); rnk >= 0 {
						req.Replies <- resource.RankedResource{Rank: rnk, Res: res}
					}
				}
			}
		}
	case ByPath:
		if res, ok := m.resources[req.Data]; ok {
			req.Replies <- resource.RankedResource{Rank: 0, Res: res}
		}
	case ByPathPrefix:
		for path, res := range m.resources {
			if strings.HasPrefix(path, req.Data) {
				req.Replies <- resource.RankedResource{Res: res}
			}
		}
	default:
		panic(fmt.Sprintf("Unknown ResourceRequestType: %d", req.ReqType))
	}
	req.Wg.Done()
}
