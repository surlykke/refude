package resource

import (
	"net/http"
	"sync"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/respond"
)

type List struct {
	sync.Mutex
	collectionPath string
	resources      []*Resource
	serveReverted  bool
}

func MakeList(collectionPath string) *List {
	return &List{
		collectionPath: collectionPath,
		resources:      make([]*Resource, 0, 20),
		serveReverted:  false,
	}
}

func MakeRevertedList(collectionPath string) *List {
	var list = MakeList(collectionPath)
	list.serveReverted = true
	return list
}

func (l *List) Get(path string) *Resource {
	l.Lock()
	defer l.Unlock()
	for _, res := range l.resources {
		if res.Path == path {
			return res
		}
	}
	return nil
}

func (l *List) GetAll() []*Resource {
	l.Lock()
	defer l.Unlock()
	var all = make([]*Resource, len(l.resources))
	copy(all, l.resources)
	return all
}

func (l *List) Put(res *Resource) {
	l.Lock()
	defer l.Unlock()
	for i := 0; i < len(l.resources); i++ {
		if l.resources[i].Path == res.Path {
			l.resources[i] = res
			return
		}
	}

	l.resources = append(l.resources, res)
}

func (l *List) ReplaceWith(resources []*Resource) {
	l.Lock()
	defer l.Unlock()
	l.resources = resources
}

func (l *List) Delete(path string) bool {
	l.Lock()
	defer l.Unlock()
	var deleted = 0
	for i := 0; i < len(l.resources); i++ {
		if l.resources[i].Path != path {
			l.resources[i-deleted] = l.resources[i]
			deleted++
		}
	}
	l.resources = l.resources[0 : len(l.resources)-deleted]
	return deleted > 0
}

func (l *List) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if l.collectionPath == r.URL.Path {
		(&Resource{
			Links: link.MakeList(r.URL.Path, "", ""),
			Title: "Collection",
			Data:  dataSlice(l.GetAll()),
		}).ServeHTTP(w, r)
	} else if res := l.Get(r.URL.Path); res != nil {
		res.ServeHTTP(w, r)
	} else {
		respond.NotFound(w)
	}
}
