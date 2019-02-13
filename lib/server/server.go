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

