// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package lib

import (
	"net/http"
	"fmt"
	"crypto/sha1"
	"encoding/json"
	"log"
	"strings"
)

// A standardized path is a path that starts with '/' and has no double slashes
type StandardizedPath string

/** transform a path to a standardized path
 * Watered down version of path.Clean. Replace any sequence of '/' with single '/'
 * Remove ending '/'
 * We do not resolve '..', (so '/foo/../baa' is different from '/baa')
 * Examples:
 *       '//foo/baa' becomes '/foo/baa'
 *       '/foo///baa/////muh/' becomes '/foo/baa/muh'
 *       '/foo/..//baa//' becomes '/foo/../baa'
 */
func Standardize(p string) StandardizedPath {
	if len(p) == 0 || p[0] != '/' {
		panic(fmt.Sprintf("path must start with '/': '%s'", p))
	}

	var buffer = make([]byte, len(p), len(p))
	var pos = 0
	var justSawSlash = false

	for i := 0; i < len(p); i++ {
		if !justSawSlash || p[i] != '/' {
			buffer[pos] = p[i]
			pos++
		}
		justSawSlash = p[i] == '/'
	}

	if buffer[pos-1] == '/' {
		return StandardizedPath(buffer[:pos-1])
	} else {
		return StandardizedPath(buffer[:pos])
	}

}

func Standardizef(format string, args...interface{}) StandardizedPath {
	return Standardize(fmt.Sprintf(format, args...))
}

/**
	Break standardized path into dir-part and base-part
    '/foo/baa/res' -> '/foo/baa', 'res'
    '/foo/baa' -> '/foo', 'baa'
 */
func separate(sp StandardizedPath) (StandardizedPath, string) {
	if len(sp) == 0 {
		panic("Separating empty string")
	} else {
		var pos = strings.LastIndexByte(string(sp[:len(sp)-1]), '/')
		return sp[:pos], string(sp[pos+1:])
	}
}



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
	GetSelf() StandardizedPath
	GetMt() MediaType
}

type AbstractResource struct {
	Self StandardizedPath `json:"_self,omitempty"`
	Relates map[MediaType][]StandardizedPath `json:"_relates,omitempty"`
	Mt MediaType `json:"-"`
}

func (ar *AbstractResource) GetSelf() StandardizedPath {
	return ar.Self
}

func (ar *AbstractResource) GetMt() MediaType {
	return ar.Mt
}

func Relate(r1, r2 *AbstractResource) {
	if r1.Self == "" || r2.Self == "" {
		log.Fatal("Relating resources with empty 'self'")
	}

	if r1.Relates == nil {
		r1.Relates = make(map[MediaType][]StandardizedPath)
	}
	if r2.Relates == nil {
		r2.Relates = make(map[MediaType][]StandardizedPath)
	}

	r1.Relates[r2.Mt] = append(r1.Relates[r2.Mt], r2.Self)
	r2.Relates[r1.Mt] = append(r2.Relates[r1.Mt], r1.Self)
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

func (jr *JsonResource) GetSelf() StandardizedPath {
	return jr.res.GetSelf()
}

func (jr *JsonResource) GetMt() MediaType {
	return jr.res.GetMt()
}

func (jr *JsonResource) Matches(matcher Matcher) bool {
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
	} else if EtagMatch(resourceEtag, etagList) {
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
