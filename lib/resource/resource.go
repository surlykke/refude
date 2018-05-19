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
	"encoding/json"
	"github.com/surlykke/RefudeServices/lib/requestutils"
	"github.com/surlykke/RefudeServices/lib/query"
	"path/filepath"
	"github.com/pkg/errors"
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
	GetSelf() string
	GetMt() mediatype.MediaType
}

type AbstractResource struct {
	Self string `json:"_self,omitempty"`
	Relates map[string]mediatype.MediaType `json:"_relates,omitempty"`
	Mt mediatype.MediaType `json:"-"`
}

func Ar(mt mediatype.MediaType) AbstractResource {
	return AbstractResource{"", make(map[string]mediatype.MediaType), mt}
}

func (ar *AbstractResource) GetSelf() string {
	return ar.Self
}

func (ar *AbstractResource) GetMt() mediatype.MediaType {
	return ar.Mt
}

func Relate(r1, r2 *AbstractResource) {
	if r1.Relates == nil {
		r1.Relates = make(map[string]mediatype.MediaType)
	}
	if r2.Relates == nil {
		r2.Relates = make(map[string]mediatype.MediaType)
	}

	if r2RelativeToR1, err := filepath.Rel(filepath.Dir(r1.Self), r2.Self); err != nil {
		panic(errors.Errorf("Cannot determine relative path from '%s' and '%s': '%s'", r1.Self, r2.Self, err.Error()))
	} else {
		r1.Relates[r2RelativeToR1] = r2.Mt
	}

	if r1RelativeToR2, err := filepath.Rel(filepath.Dir(r2.Self), r1.Self); err != nil {
		panic(errors.Errorf("Cannot determine relative path from '%s' and '%s': '%s'", r2.Self, r1.Self, err.Error()))
	} else {
		r2.Relates[r1RelativeToR2] = r1.Mt
	}
}



type JsonResource struct {
	res          Resource
	data         []byte
	etag         string
	readyToServe bool
}

func (jr *JsonResource) GetRes() Resource {
	return jr.res
}

func (jr *JsonResource) GetSelf() string {
	return jr.res.GetSelf()
}

func (jr *JsonResource) GetMt() mediatype.MediaType {
	return jr.res.GetMt()
}

func (jr *JsonResource) Matches(matcher query.Matcher) bool {
	return matcher(jr.res)
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
		fmt.Println("JsonResource doing POST")
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
	} else if requestutils.EtagMatch(resourceEtag, etagList) {
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
		panic("Could not json-marshal")
	} else {
		return bytes
	}
}

func MakeJsonResource(res Resource) *JsonResource {
	return &JsonResource{res: res}
}
