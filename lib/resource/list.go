package resource

import (
	"net/http"
	"sync"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

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
	var resource Resource
	var found = false
	l.Lock()
	defer l.Unlock()

	if l.collectionPath != "" && r.URL.Path == l.collectionPath {
		resource, found = Resource{
			Links: link.MakeList(l.collectionPath, "", ""),
			Path:  l.collectionPath,
			Title: "Collection",
			Data:  dataSlice(l.getAll()),
		}, true
	} else {
		for _, res := range l.resources {
			if res.Path == r.URL.Path {
				resource, found = res, true
				break
			}
		}
	}

	if found {
		resource.ServeHTTP(w, r)
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

func (l *List) MakeAndPut(path, title, comment, iconName string, data Data) {
	l.Put(MakeResource(path, title, comment, iconName, l.profile, data))
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
	var newResources = make([]Resource, 0, len(l.resources))
	var found = false
	for _, res := range l.resources {
		if res.Path != path {
			newResources = append(newResources, res)
		} else {
			found = true
		}
	}
	l.resources = newResources
	return found
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

func (l *List) find(test func(res Resource) bool) []Resource {
	var result = make([]Resource, 0, 3)
	l.Lock()
	defer l.Unlock()

	for _, res := range l.resources {
		if test(res) {
			result = append(result, res)
		}
	}
	return result
}
