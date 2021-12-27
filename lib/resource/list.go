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
}

func MakeList(collectionPath string) *List {
	return &List{
		collectionPath: collectionPath,
		resources:      make([]*Resource, 0, 20),
	}
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

func (l *List) PutFirst(res *Resource) {
	l.Lock()
	defer l.Unlock()
	for i := 0; i < len(l.resources); i++ {
		if l.resources[i].Path == res.Path {
			l.resources[i] = res
			return
		}
	}

	l.resources = append([]*Resource{res}, l.resources...)
}

func (l *List) ReplaceWith(resources []*Resource) {
	l.Lock()
	defer l.Unlock()
	l.resources = resources
}

func (l *List) Delete(path string) bool {
	l.Lock()
	defer l.Unlock()

	var retained = 0
	for i := 0; i < len(l.resources); i++ {
		if l.resources[i].Path != path {
			l.resources[retained] = l.resources[i]
			retained++
		}

	}
	if retained < len(l.resources) {
		l.resources = l.resources[:retained]
		return true
	} else {
		return false
	}

}

func (l *List) FindFirst(test func(data Data) bool) Data {
	l.Lock()
	defer l.Unlock()
	for _, resource := range l.resources {
		if test(resource.Data) {
			return resource.Data
		}
	}
	return nil
}

func (l *List) DeleteIf(cond func(res *Resource) bool) {
	l.Lock()
	defer l.Unlock()
	var retained = 0
	for i := 0; i < len(l.resources); i++ {
		if !cond(l.resources[i]) {
			l.resources[retained] = l.resources[i]
			retained++
		}
	}
	if retained < len(l.resources) {
		l.resources = l.resources[:retained]
	}
}

func (l *List) Walk(walker func(res *Resource)) {
	l.Lock()
	defer l.Unlock()
	for _, res := range l.resources {
		walker(res)
	}
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
