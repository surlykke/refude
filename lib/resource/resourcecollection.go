package resource

import (
	"net/http"
)

type ResourceCollection interface {
	OwnsPath(path string) bool
	Get(path string) Resource
	GetByPrefix(prefix string) []Resource
}

func ServeHttp(rc ResourceCollection, w http.ResponseWriter, r *http.Request) bool {
	if !rc.OwnsPath(r.URL.Path) {
		return false
	}

	if rl := rc.GetByPrefix(r.URL.Path); rl != nil {
		ServeCollection(w, r, rl)
	} else if resource := rc.Get(r.URL.Path); resource != nil {
		ServeResource(w, r, resource)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

	return true
}
