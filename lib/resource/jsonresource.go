package resource

import (
	"crypto/sha1"
	"fmt"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/requests"
)

type JsonResource struct {
	bytes []byte
	etag  string
	data  interface{}
}

func MakeJsonResource(data interface{}) *JsonResource {
	return &JsonResource{bytes: ToJson(data), data: data}
}

func MakeJsonResouceWithEtag(data interface{}) *JsonResource {
	var jsonResource = MakeJsonResource(data)
	jsonResource.etag = fmt.Sprintf("\"%x\"", sha1.Sum(jsonResource.bytes))
	return jsonResource
}

func (jr *JsonResource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if statusCode := requests.CheckEtag(r, jr.etag); statusCode != 0 {
			w.WriteHeader(statusCode)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("ETag", jr.etag)
			_, _ = w.Write(jr.bytes)
		}
	} else if posthandler, ok := jr.data.(PostHandler); ok && r.Method == "POST" {
		posthandler.POST(w, r)
	} else if patchHandler, ok := jr.data.(PatchHandler); ok && r.Method == "PATCH" {
		patchHandler.PATCH(w, r)
	} else if deleteHandler, ok := jr.data.(DeleteHandler); ok && r.Method == "DELETE" {
		deleteHandler.DELETE(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func (jr *JsonResource) GetEtag() string {
	return jr.etag
}
