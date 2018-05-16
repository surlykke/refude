// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"net/http"
	"fmt"
	"crypto/sha1"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/query"
	"encoding/json"
)

type GetHandler interface {
	GET(w http.ResponseWriter, r *http.Request)
}

type PostHandler interface {
	POST(w http.ResponseWriter, r *http.Request)
}

type PatchHandler interface {
	PATCH(w http.ResponseWriter, r *http.Request)
}

type DeleteHandler interface {
	DELETE(w http.ResponseWriter, r *http.Request)
}

type Resource interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Mt() mediatype.MediaType
	Match(m query.Matcher) bool
}

type MappableType interface {
	SetSelf(string)
}

type Self struct {
	Self string `json:"_self,omitempty"`
}

func (s *Self) SetSelf(path string) {
	s.Self = path
}

type JsonResource struct {
	Res       interface{}
	data      []byte
	mediaType mediatype.MediaType
	etag      string
	prepared  bool
}

// Caller must make sure that no other goroutine accesses during this.
func (jr *JsonResource) Prepare() {
	if !jr.prepared {
		jr.data = ToJSon(jr.Res)
		jr.etag = fmt.Sprintf("\"%x\"", sha1.Sum(jr.data))
		jr.prepared = true
	}
}

func (jr *JsonResource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Header().Set("Content-Type", string(jr.mediaType))
		w.Header().Set("ETag", jr.etag)
		w.Write(jr.data)
		return
	case "POST":
		fmt.Println("JsonResource doing POST")
		if postHandler, ok := jr.Res.(PostHandler); ok {
			fmt.Println("Found handler")
			postHandler.POST(w, r)
			return
		}
	case "PATCH":
		if patchHandler, ok := jr.Res.(PatchHandler); ok {
			patchHandler.PATCH(w, r)
			return
		}
	case "DELETE":
		if deleteHandler, ok := jr.Res.(DeleteHandler); ok {
			deleteHandler.DELETE(w, r)
			return
		}
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (jr *JsonResource) MarshalJSON() ([]byte, error) {
	return jr.data, nil
}

func (jr *JsonResource) Mt() mediatype.MediaType {
	return jr.mediaType
}

func ToJSon(res interface{}) []byte {
	if bytes, err := json.Marshal(res); err != nil {
		panic("Could not json-marshal")
	} else {
		return bytes
	}
}

func MakeJsonResource(res interface{}, mediaType mediatype.MediaType) *JsonResource {
	return &JsonResource{Res: res, mediaType:mediaType}
}
