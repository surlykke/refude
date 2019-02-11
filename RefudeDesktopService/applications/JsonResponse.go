package applications

import (
	"crypto/sha1"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"net/http"
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

func (jr *JsonResponse) GET(w http.ResponseWriter) {

}

type ResourceCollection interface {
	GetResource(r *http.Request) (interface{}, error)
}

type JsonResponseCache struct {
	jsonResponses map[string]*JsonResponse
	resources ResourceCollection
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


type Server struct {
	resources ResourceCollection
	cache JsonResponseCache
	lock sync.Mutex
}

func MakeServer(resources ResourceCollection) *Server {
	return &Server{
		resources: resources,
		cache: MakeJsonResponseCache(resources),
	}
}

func (s *Server) getJsonResponse(r *http.Request) *JsonResponse {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.cache.GetJsonResponse(r)
}

func (s *Server) getResource(r *http.Request) (interface{}, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.resources.GetResource(r)
}


func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if "GET" == r.Method {
		if jsonResponse := s.getJsonResponse(r); jsonResponse == nil {
			w.WriteHeader(http.StatusNotFound)
		} else if jsonResponse.error != nil {
			requests.ReportUnprocessableEntity(w, jsonResponse.error)
		} else {
			w.Header().Set("Content-Type", string(jsonResponse.ContentType))
			w.Header().Set("ETag", jsonResponse.Etag)
			w.Write(jsonResponse.Data)
		}
	} else {
		if res, err := s.getResource(r); err != nil {
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

func (s *Server) setResources(resources ResourceCollection) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.resources = resources
	s.cache = MakeJsonResponseCache(resources)
}


