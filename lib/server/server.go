package server

import (
	"crypto/sha1"
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

func MakeJsonResponse(res interface{}, ContentType resource.MediaType, error error) *JsonResponse {
	var jsonResponse = &JsonResponse{}
	if error == nil {
		jsonResponse.Data = resource.ToJSon(res)
		jsonResponse.Etag = fmt.Sprintf("\"%x\"", sha1.Sum(jsonResponse.Data))
		jsonResponse.ContentType = ContentType
	} else {
		jsonResponse.error = error
	}
	return jsonResponse
}


type ResourceCollection interface {
	sync.Locker
	GetJsonResponse(r *http.Request) *JsonResponse
	GetResource(r *http.Request) (interface{}, error)
}

type ResourceCollection2 interface {
	GetResource(r *http.Request) (interface{}, error)
}

type ResourceServer interface {
	resource.GetHandler
	resource.PostHandler
	resource.PatchHandler
	resource.DeleteHandler
	HandledPrefixes() []string
}

type PostNotAllowed struct {}

func (PostNotAllowed) POST(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type PatchNotAllowed struct {}

func (PatchNotAllowed) PATCH(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type DeleteNotAllowed struct {}

func (DeleteNotAllowed) DELETE(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}




type JsonResponseCache struct {
	jsonResponses map[string]*JsonResponse
	resources ResourceCollection
}

type JsonResponseCache2 struct {
	cachedResponses map[string]*JsonResponse
	mutex sync.Mutex
	resources ResourceCollection2
}


func MakeJsonResponseCache2(resources ResourceCollection2) JsonResponseCache2 {
	return JsonResponseCache2{cachedResponses: make(map[string]*JsonResponse), resources: resources}
}

func (jrc *JsonResponseCache2) getCachedResponse(path string) (*JsonResponse, bool) {
	jrc.mutex.Lock()
	defer jrc.mutex.Unlock()
	resp, ok := jrc.cachedResponses[path]
	return resp,ok
}

func (jrc *JsonResponseCache2) setCachedResponse(path string, response *JsonResponse) {
	jrc.mutex.Lock()
	defer jrc.mutex.Unlock()
	jrc.cachedResponses[path] = response
}


func (jrc JsonResponseCache2) ClearByPrefixes(prefixes ...string) {
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

func (jrc JsonResponseCache2) Clear() {
	jrc.mutex.Lock()
	defer jrc.mutex.Unlock()

	jrc.cachedResponses = make(map[string]*JsonResponse)
}


func (jrc *JsonResponseCache2) GET(w http.ResponseWriter, r *http.Request) {
	response, ok := jrc.getCachedResponse(r.URL.RequestURI());

	if (!ok) {
		if res, err := jrc.resources.GetResource(r); err != nil {
			response = MakeJsonResponse(nil, "", err)
		} else if res != nil {
			response = MakeJsonResponse(res, "application/json", nil)
		}

		jrc.setCachedResponse(r.URL.RequestURI(), response)
	}

	if response == nil {
		w.WriteHeader(http.StatusNotFound)
	} else if response.error != nil {
		requests.ReportUnprocessableEntity(w, response.error)
	} else {
		w.Header().Set("Content-Type", string(response.ContentType))
		w.Header().Set("ETag", response.Etag)
		w.Write(response.Data)

	}
}



func MakeJsonResponseCache(resources ResourceCollection) JsonResponseCache {
	return JsonResponseCache{
		make(map[string]*JsonResponse),
		resources,
	}
}

func (jrc JsonResponseCache) GetJsonResponse(r *http.Request) *JsonResponse {
	jsonResponse, ok := jrc.jsonResponses[r.RequestURI]

	if (!ok) {
		if res, err := jrc.resources.GetResource(r); err != nil {
			jsonResponse = MakeJsonResponse(nil, "", err)
		} else if res != nil {
			jsonResponse = MakeJsonResponse(res, "application/json", nil)
		}

		jrc.jsonResponses[r.RequestURI] = jsonResponse
	}

	return jsonResponse
}

func (jrc JsonResponseCache) ClearByPrefixes(prefixes ...string) {
	for path, _ := range jrc.jsonResponses {
		for _, prefix := range prefixes {
			if strings.HasPrefix(path, prefix) {
				delete(jrc.jsonResponses, path)
				break
			}
		}
	}
}

func (jrc JsonResponseCache) Clear() {
	jrc.jsonResponses = make(map[string]*JsonResponse)
}


type Server struct {
	resources ResourceCollection
}

func MakeServer(resources ResourceCollection) Server {
	return Server{resources}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if "GET" == r.Method {
		s.resources.Lock()
		var jsonResponse = s.resources.GetJsonResponse(r)
		s.resources.Unlock()

		if jsonResponse == nil {
			w.WriteHeader(http.StatusNotFound)
		} else if jsonResponse.error != nil {
			requests.ReportUnprocessableEntity(w, jsonResponse.error)
		} else {
			w.Header().Set("Content-Type", string(jsonResponse.ContentType))
			w.Header().Set("ETag", jsonResponse.Etag)
			w.Write(jsonResponse.Data)
		}
	} else {
		s.resources.Lock()
		res, err := s.resources.GetResource(r)
		s.resources.Unlock()

		if err != nil {
			requests.ReportUnprocessableEntity(w, err)
		} else if res == nil {
			w.WriteHeader(http.StatusNotFound)
		} else {
			if postHandler, ok := res.(resource.PostHandler); ok && "POST" == r.Method {
				postHandler.POST(w, r)
			} else if patchHandler, ok := res.(resource.PatchHandler); ok && "PATCH" == r.Method {
				patchHandler.PATCH(w, r)
			} else if deleteHandler, ok := res.(resource.DeleteHandler); ok && "DELETE" == r.Method {
				deleteHandler.DELETE(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed);
			}
		}
	}
}

