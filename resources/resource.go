package resources

import (
	"net/http"
	"sync"
	"encoding/json"
)

type FallbackHandler struct {}

func (fb FallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type JsonResource struct {
	data  http.Handler
	bytes []byte
	mutex sync.RWMutex
}

func NewJsonResource(handler http.Handler) JsonResource {
	return JsonResource{handler, nil, sync.RWMutex{}}
}

func (jr JsonResource) ServeHTTP(w http.ResponseWriter, r *http.Request){
	jr.mutex.RLock()
	defer jr.mutex.RUnlock()
	if jr.bytes == nil {
		if !jr.getBytes() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.Write(jr.bytes)
}

func (jr *JsonResource) getBytes() bool {
	jr.mutex.RUnlock()
	defer jr.mutex.RLock()

	jr.mutex.Lock()
	defer jr.mutex.Unlock()

	var err error
	jr.bytes, err = json.Marshal(jr.data)

	return err == nil
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

func (rc* ResourceCollection) Update(resources map[string]http.Handler) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	for path,_ := range rc.resources {
		if _,isThere := resources[path]; !isThere {
			rc.notifier.Notify("resource-removed", path)
		}
	}

	for path,handler := range resources {
		if oldHandler, isThere := rc.resources[path]; isThere {
			if oldHandler != handler {
				rc.notifier.Notify("resource-updated", path)
			}
		} else {
			rc.notifier.Notify("resource-added", path)
		}
	}

	rc.resources = resources
}
