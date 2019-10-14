package resource

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/surlykke/RefudeServices/lib/requests"
)

type (
	Resource interface {
		GetSelf() string
		POST(w http.ResponseWriter, r *http.Request)
		PATCH(w http.ResponseWriter, r *http.Request)
		DELETE(w http.ResponseWriter, r *http.Request)
	}

	GetHandler interface {
		GET(w http.ResponseWriter, r *http.Request)
	}
)

var (
	/** Deadlock prevention:
	 * condsLock and resourcesLock may not be held at the same time
	 * only one conds[...].L may be held at a time
	 * When a conds[...].L and resourcesLock are to be held conds[...].L must be taken first.
	 */

	resouces = make(map[string]Resource)

	collections   = make(map[string]map[string]bool)
	resources     = make(map[string]Resource)
	resourcesLock sync.Mutex

	conds     = make(map[string]*sync.Cond)
	condsLock sync.Mutex
)

func get(path string) (Resource, bool) {
	resourcesLock.Lock()
	var res, ok = resources[path]
	resourcesLock.Unlock()
	return res, ok
}

func getCond(path string) (*sync.Cond, bool) {
	condsLock.Lock()
	cond, ok := conds[path]
	condsLock.Unlock()
	return cond, ok
}

func MapSingle(path string, res Resource) {
	resourcesLock.Lock()
	resources[path] = res
	resourcesLock.Unlock()

	condsLock.Lock()
	if cond, ok := conds[path]; ok {
		cond.Broadcast()
	}
	condsLock.Unlock()
}

/**
 * Allows to map a collection of resources, giving that collection a name. Next time a collection
 * is mapped with that name it fully replaces the first collection
 */
func MapCollection(resourcesToMap *map[string]Resource, collectionName string) {
	var affectedPaths = make(map[string]bool)
	resourcesLock.Lock()
	if oldCollection, ok := collections[collectionName]; ok {
		for path, _ := range oldCollection {
			delete(resources, path)
			affectedPaths[path] = true
		}
	}

	var newCollection = make(map[string]bool)

	for path, resource := range *resourcesToMap {
		resources[path] = resource
		newCollection[path] = true
		affectedPaths[path] = true
	}

	collections[collectionName] = newCollection
	resourcesLock.Unlock()

	condsLock.Lock()
	for path, _ := range affectedPaths {
		if cond, ok := conds[path]; ok {
			cond.Broadcast()
		}
	}
	condsLock.Unlock()
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if requests.HaveParam(r, "longpoll") {
		longGet(w, r)
	} else {
		if resource, ok := get(r.URL.Path); !ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			switch r.Method {
			case "GET":
				if getHandler, ok := resource.(GetHandler); ok {
					getHandler.GET(w, r)
				} else {
					ServeAsJson(w, r, resource)
				}
			case "POST":
				resource.POST(w, r)
			case "PATCH":
				resource.PATCH(w, r)
			case "DELETE":
				resource.DELETE(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}
	}
}

func longGet(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Doing longpoll")
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var path = r.URL.Path
	var (
		cond *sync.Cond
		ok   bool
	)
	condsLock.Lock()
	if cond, ok = conds[path]; !ok {
		cond = sync.NewCond(&sync.Mutex{})
		conds[path] = cond
	}
	condsLock.Unlock()

	fmt.Println("cond:", cond)
	cond.L.Lock()
	for {
		if res, ok := get(r.URL.Path); !ok {
			cond.L.Unlock()
			w.WriteHeader(http.StatusNotFound)
			return
		} else if getHandler, ok := res.(GetHandler); ok {
			cond.L.Unlock()
			getHandler.GET(w, r)
			return
		} else {
			var bytes, etag = ToJsonAndEtag(res)

			fmt.Println("Compare", etag, "to", r.Header.Get("If-None-Match"))
			if etag == "" || !requests.EtagMatch(etag, r.Header.Get("If-None-Match")) {
				fmt.Println("No match..")
				cond.L.Unlock()
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("ETag", etag)
				_, _ = w.Write(bytes)
				return
			}
		}

		cond.Wait()
	}
}

func ToJsonAndEtag(res interface{}) ([]byte, string) {
	var bytes, err = json.Marshal(res)
	if err != nil {
		panic(fmt.Sprintln(err))
	}
	return bytes, fmt.Sprintf("\"%x\"", sha1.Sum(bytes))
}

func ServeAsJson(w http.ResponseWriter, r *http.Request, res interface{}) {
	var jsonBytes, etag = ToJsonAndEtag(res)

	if statusCode := requests.CheckEtag(r, etag); statusCode != 0 {
		w.WriteHeader(statusCode)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", etag)
		_, _ = w.Write(jsonBytes)
	}

}
