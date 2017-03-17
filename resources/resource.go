package resources

import (
	"net/http"
	"sync"
	"encoding/json"
)

type Pathlist []string

func (pl Pathlist) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type FallbackHandler struct {}

func (fb FallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type JsonResource struct {
	data  http.Handler
	bytes []byte
}

func NewJsonResource(handler http.Handler) JsonResource {
	if byte, err := json.Marshal(handler); err == nil {
		return JsonResource{handler, byte}
	} else {
		panic(err)
	}

}

func (jr JsonResource) ServeHTTP(w http.ResponseWriter, r *http.Request){
	if r.Method != "GET" {
		jr.data.ServeHTTP(w, r)
	} else {
		w.Write(jr.bytes)
	}
}

type ResourceCollection struct {
	resources map[string]http.Handler
	notifier Notifier
	mutex sync.RWMutex
}

func NewResourceCollection() ResourceCollection {
	return ResourceCollection{make(map[string]http.Handler), NewNotifier(), sync.RWMutex{}}
}

func (rc ResourceCollection) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	if r.URL.Path == "/notifications" {
		rc.notifier.ServeHTTP(w, r)
	} else if handler, ok := rc.resources[r.URL.Path]; ok {
		handler.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (rc* ResourceCollection) Set(resources map[string]http.Handler) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	for path,handler := range resources {
		if oldHandler, isThere := rc.resources[path]; isThere {
			if oldHandler != handler {
				rc.notifier.Notify("resource-updated", path)
			}
		} else {
			rc.notifier.Notify("resource-added", path)
		}

		rc.resources[path] = handler
	}
}

func (rc* ResourceCollection) Remove(paths []string) {
	for _, path := range paths {
		if _,ok := rc.resources[path]; ok {
			rc.notifier.Notify("resource-removed", path)
			delete(rc.resources, path)
		}
	}
}

