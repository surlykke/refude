package resource

import (
	"net/http"
	"sync"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

type Data interface {
	ForDisplay() bool
	Links(path string) link.List
}

type Resource struct {
	Links   link.List `json:"_links"`
	Path    string    `json:"-"`
	Title   string    `json:"title"`
	Comment string    `json:"comment,omitempty"`
	Icon    link.Href `json:"icon,omitempty"`
	Profile string    `json:"profile"`
	Data    Data      `json:"data"`
}

func Make(path, title, comment, iconName, profile string, data Data) Resource {
	return Resource{
		Links:   append(link.List{link.Make(path, "", "", relation.Self)}, data.Links(path)...),
		Path:    path,
		Title:   title,
		Comment: comment,
		Icon:    link.IconUrl(iconName),
		Profile: profile,
		Data:    data,
	}
}

type Postable interface {
	DoPost(w http.ResponseWriter, r *http.Request)
}

type Deleteable interface {
	DoDelete(w http.ResponseWriter, r *http.Request)
}

func (res Resource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		respond.AsJson(w, res)
		return
	case "POST":
		if postable, ok := res.Data.(Postable); ok {
			postable.DoPost(w, r)
			return
		}
	case "DELETE":
		if deleteable, ok := res.Data.(Deleteable); ok {
			deleteable.DoDelete(w, r)
			return
		}
	}
	respond.NotAllowed(w)

}

type dataSlice []Resource

func (ds dataSlice) Links(path string) link.List {
	return link.List{}
}

func (ds dataSlice) ForDisplay() bool {
	return false
}

type List struct {
	sync.Mutex
	profile        string
	insertAtFront  bool
	collectionPath string
	alternatePaths map[string]string
	resources      []Resource
}

func MakeList(profile string, insertAtFront bool, collectionPath string, initialCap int) *List {
	return &List{
		profile:        profile,
		insertAtFront:  insertAtFront,
		collectionPath: collectionPath,
		alternatePaths: make(map[string]string),
		resources:      make([]Resource, 0, initialCap),
	}
}

func (l *List) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if res, ok := l.Get(r); ok {
		res.ServeHTTP(w, r)
	} else {
		respond.NotFound(w)
	}
}

func (l *List) Put(res Resource) {
	l.Lock()
	defer l.Unlock()

	for i := 0; i < len(l.resources); i++ {
		if l.resources[i].Path == res.Path {
			l.resources[i] = res
			return
		}
	}

	if l.insertAtFront {
		l.resources = append([]Resource{res}, l.resources...)
	} else {
		l.resources = append(l.resources, res)
	}

}

func (l *List) Put2(path, title, comment, iconName string, data Data) {
	l.Put(Make(path, title, comment, iconName, l.profile, data))
}

func (l *List) GetData(path string) Data {
	for _, res := range l.resources {
		if res.Path == path {
			return res.Data
		}
	}

	return nil
}

func (l *List) Get(r *http.Request) (Resource, bool) {
	if l.collectionPath != "" && r.URL.Path == l.collectionPath {
		return Resource{
			Links: link.MakeList(l.collectionPath, "", ""),
			Path:  l.collectionPath,
			Title: "Collection",
			Data:  dataSlice(l.getAll()),
		}, true
	} else {
		for i := 0; i < len(l.resources); i++ {
			if l.resources[i].Path == r.URL.Path {
				var resCopy = l.resources[i]
				if r.Method == "GET" {
					resCopy.Links = resCopy.Links.Filter(requests.Term(r))
				}
				return resCopy, true
			}
		}
	}
	return Resource{}, false
}

func (l *List) Delete(path string) bool {
	l.Lock()
	defer l.Unlock()
	for i := 0; i < len(l.resources); i++ {
		if l.resources[i].Path == path {
			copy(l.resources[:i], l.resources[:i+1])
			l.resources = l.resources[:len(l.resources)-1]
			return true
		}
	}

	return false
}

func (l *List) ReplaceWith(resources []Resource) {
	l.Lock()
	defer l.Unlock()
	var n = len(resources)
	l.resources = make([]Resource, n)
	if l.insertAtFront {
		for i := 0; i < n; i++ {
			l.resources[n-i-1] = resources[i]
		}
	} else {
		copy(l.resources, resources)
	}
}

func (l *List) GetAll() []Resource {
	l.Lock()
	defer l.Unlock()
	return l.getAll()
}

func (l *List) getAll() []Resource {
	var list = make([]Resource, len(l.resources), len(l.resources))
	copy(list, l.resources)
	return list
}

func (l *List) Collection(path, title string) Resource {
	var resources = make(dataSlice, len(l.resources))
	copy(resources, l.resources)
	var collection = Resource{
		Links: link.MakeList(path, title, ""),
		Path:  path,
		Title: title,
		Data:  resources,
	}
	return collection
}

func (l *List) Search(term string, sink chan link.Link) {
	for i, res := range l.GetAll() {
		if res.Data.ForDisplay() {
			var rnk int
			if term == "" {
				rnk = i
			} else {
				rnk = searchutils.Match(term, res.Title)
			}
			if rnk > -1 {
				sink <- link.MakeRanked2(res.Path, res.Title, res.Icon, res.Profile, rnk)
			}
		}
	}
}

func (l *List) Paths(method string, sink chan string) {
	l.Lock()
	defer l.Unlock()

	for _, res := range l.resources {
		sink <- res.Path
	}
}
