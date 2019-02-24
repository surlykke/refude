package server

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"net/http"
	"strings"
	"sync"
)

type JsonResponse struct {
	Data        []byte
	Etag        string
	ContentType resource.MediaType
	error       error
}

func ToJSon(res interface{}) []byte {
	if bytes, err := json.Marshal(res); err != nil {
		panic(fmt.Sprintln(err))
		return nil
	} else {
		return bytes
	}
}

func MakeJsonResponse(res interface{}, ContentType resource.MediaType, error error) *JsonResponse {
	var jsonResponse = &JsonResponse{}
	if error == nil {
		jsonResponse.Data = ToJSon(res)
		jsonResponse.Etag = fmt.Sprintf("\"%x\"", sha1.Sum(jsonResponse.Data))
		jsonResponse.ContentType = ContentType
	} else {
		jsonResponse.error = error
	}
	return jsonResponse
}

type ResourceCollection interface {
	GetSingle(r *http.Request) interface{}       // nil means not found
	GetCollection(r *http.Request) []interface{} // nil means not found, empty slice means found (but, well, empty)
}

type ResourceServer interface {
	resource.GetHandler
	resource.PostHandler
	resource.PatchHandler
	resource.DeleteHandler
	HandledPrefixes() []string
}

type PostNotAllowed struct{}

func (PostNotAllowed) POST(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type PatchNotAllowed struct{}

func (PatchNotAllowed) PATCH(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type DeleteNotAllowed struct{}

func (DeleteNotAllowed) DELETE(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type CachingJsonGetter struct {
	cachedResponses map[string]*JsonResponse
	mutex           sync.Mutex
	resources       ResourceCollection
}

func MakeCachingJsonGetter(resources ResourceCollection) CachingJsonGetter {
	return CachingJsonGetter{cachedResponses: make(map[string]*JsonResponse), resources: resources}
}

func (cjg *CachingJsonGetter) GET(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CachingJsonGetter GET")
	cjg.mutex.Lock()
	var response, ok = cjg.cachedResponses[r.RequestURI]
	cjg.mutex.Unlock()

	if !ok {
		fmt.Println("Not in cache, query:", r.URL.Query())
		if collection := cjg.resources.GetCollection(r); collection != nil {
			fmt.Println("collection")
			var matcher, err = requests.GetMatcher2(r);
			fmt.Println("matcher, err:", matcher, err)
			if err != nil {
				response = MakeJsonResponse(nil, "", err)
			} else if matcher != nil {
				fmt.Println("Filter")
				var filtered = make([]interface{}, 0, len(collection))
				for _, res := range collection {
					if matcher(res) {
						filtered = append(filtered, res)
					}
				}
				response = MakeJsonResponse(filtered, "application/json", nil)
			} else {
				response = MakeJsonResponse(collection, "application/json", nil)
			}
		} else if res := cjg.resources.GetSingle(r); res != nil {
			fmt.Println("Single")
			response = MakeJsonResponse(res, "application/json", nil)
		} else {
			fmt.Println("nil")
			response = nil
		}

		cjg.mutex.Lock()
		cjg.cachedResponses[r.RequestURI] = response
		cjg.mutex.Unlock()
	} else {
		fmt.Println("In cache")
	}

	if response == nil {
		w.WriteHeader(http.StatusNotFound)
	} else if response.error != nil {
		requests.ReportUnprocessableEntity(w, response.error)
	} else {
		w.Header().Set("Content-Type", string(response.ContentType))
		w.Header().Set("ETag", response.Etag)
		_,_ = w.Write(response.Data)

	}

}

func (jrc CachingJsonGetter) ClearByPrefixes(prefixes ...string) {
	jrc.mutex.Lock()
	defer jrc.mutex.Unlock()

	for path, _ := range jrc.cachedResponses {
		for _, prefix := range prefixes {
			if strings.HasPrefix(path, prefix) {
				delete(jrc.cachedResponses, path)
				break
			}
		}
	}
}

func (jrc CachingJsonGetter) Clear() {
	jrc.mutex.Lock()
	defer jrc.mutex.Unlock()

	jrc.cachedResponses = make(map[string]*JsonResponse)
}



