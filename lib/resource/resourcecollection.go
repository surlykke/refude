package resource

import (
	"net/http"
)

type ResourceCollection interface {
	Get(path string) Resource
	GetList(path string) []Resource
}

func ServeHttp(rc ResourceCollection, w http.ResponseWriter, r *http.Request) bool {
	if rl := rc.GetList(r.URL.Path); rl != nil {
		ServeCollection(w, r, rl)
	} else if resource := rc.Get(r.URL.Path); resource != nil {
		ServeResource(w, r, resource)
	} else {
		return false
	}

	return true
}
