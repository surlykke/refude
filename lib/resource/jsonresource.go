package resource

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/requests"
	"net/http"
)

type JsonResource struct {
	res          Resource
	data         []byte
	etag         string
	readyToServe bool
}

func (jr *JsonResource) GetRes() Resource {
	return jr.res
}

func (jr *JsonResource) GetSelf() StandardizedPath {
	return jr.res.GetSelf()
}

func (jr *JsonResource) GetMt() MediaType {
	return jr.res.GetMt()
}

// Caller must make sure that no other goroutine accesses during this.
func (jr *JsonResource) EnsureReady() {
	if !jr.readyToServe {
		jr.data = ToJSon(jr.res)
		jr.etag = fmt.Sprintf("\"%x\"", sha1.Sum(jr.data))
		jr.readyToServe = true
	}
}

func (jr *JsonResource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if preventedByEtagCondition(r, jr.etag, true) {
			w.WriteHeader(http.StatusNotModified)
		} else {
			w.Header().Set("Content-Type", string(jr.res.GetMt()))
			w.Header().Set("ETag", jr.etag)
			w.Write(jr.data)
		}
		return
	case "POST":
		if postHandler, ok := jr.res.(PostHandler); ok {
			if preventedByEtagCondition(r, jr.etag, false) {
				w.WriteHeader(http.StatusPreconditionFailed)
			} else {
				postHandler.POST(w, r)
			}
			return
		}
	case "PATCH":
		if patchHandler, ok := jr.res.(PatchHandler); ok {
			if preventedByEtagCondition(r, jr.etag, false) {
				w.WriteHeader(http.StatusPreconditionFailed)
			} else {
				patchHandler.PATCH(w, r)
			}
			return
		}
	case "DELETE":
		if deleteHandler, ok := jr.res.(DeleteHandler); ok {
			if preventedByEtagCondition(r, jr.etag, false) {
				w.WriteHeader(http.StatusPreconditionFailed)
			} else {
				deleteHandler.DELETE(w, r)
			}
			return
		}
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func preventedByEtagCondition(r *http.Request, resourceEtag string, safeMethod bool) bool {
	var etagList string
	if safeMethod { // Safe methods are GET and HEAD
		etagList  = r.Header.Get("If-None-Match")
	} else {
		etagList = r.Header.Get("If-Match")
	}

	if etagList == "" {
		return false
	} else if requests.EtagMatch(resourceEtag, etagList) {
		return safeMethod
	} else {
		return !safeMethod
	}
}

func (jr *JsonResource) MarshalJSON() ([]byte, error) {
	return jr.data, nil
}

func ToJSon(res interface{}) []byte {
	if bytes, err := json.Marshal(res); err != nil {
		panic(fmt.Sprintln(err))
		return nil
	} else {
		return bytes
	}
}

func MakeJsonResource(res Resource) *JsonResource {
	return &JsonResource{res: res}
}
