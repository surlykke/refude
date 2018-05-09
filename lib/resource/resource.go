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
	Mt()   mediatype.MediaType
	Match(m query.Matcher) bool
}

type JsonResource struct {
	res       interface{}
	data 	  []byte
	mediaType mediatype.MediaType
	etag      string
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
		if postHandler, ok := jr.res.(PostHandler); ok {
			fmt.Println("Found handler")
			postHandler.POST(w, r)
			return
		}
	case "PATCH":
		if patchHandler, ok := jr.res.(PatchHandler); ok {
			patchHandler.PATCH(w, r)
			return
		}
	case "DELETE":
		if deleteHandler, ok := jr.res.(DeleteHandler); ok {
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

func (jr *JsonResource) Match(m query.Matcher) bool {
	return m(jr.res)
}

func MakeJsonResource(res interface{}, mediaType mediatype.MediaType) *JsonResource {
	var data = mediatype.ToJSon(res)
	var etag = fmt.Sprintf("\"%x\"", sha1.Sum(data))
	return &JsonResource{res, data, mediaType, etag}
}
