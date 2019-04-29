// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/requests"
)

type MediaType string

type Relation string

const (
	Self               Relation = "self"
	Related                     = "related"
	Associated                  = "http://relations.refude.org/associated"
	DefaultApplication          = "http://relations.refude.org/default_application"
	SNI_MENU                    = "http://relations.refude.org/sni_menu"
)

type Resource interface {
	GetSelf() string
	GetMt() MediaType
	GetEtag() string
	POST(w http.ResponseWriter, r *http.Request)
	PATCH(w http.ResponseWriter, r *http.Request)
	DELETE(w http.ResponseWriter, r *http.Request)
}

type Link struct {
	Href  string    `json:"href"`
	Rel   Relation  `json:"rel"` // We never have more than one relation on a link - we'll make a new link with same href
	Type  MediaType `json:",omitempty"`
	Title string    `json:",omitempty"`
}

type ResourceList []Resource

// For sorting
func (rc ResourceList) Len() int           { return len(rc) }
func (rc ResourceList) Swap(i, j int)      { rc[i], rc[j] = rc[j], rc[i] }
func (rc ResourceList) Less(i, j int) bool { return rc[i].GetSelf() < rc[j].GetSelf() }

type JsonResponse struct {
	Data []byte
	Etag string
}

func ToJSon(res interface{}) JsonResponse {
	var jsonResponse = JsonResponse{}
	if bytes, err := json.Marshal(res); err != nil {
		panic(fmt.Sprintln(err))
	} else {
		jsonResponse.Data = bytes
		jsonResponse.Etag = fmt.Sprintf("\"%x\"", sha1.Sum(jsonResponse.Data))
	}
	return jsonResponse

}

func ServeCollection(w http.ResponseWriter, r *http.Request, collection []Resource) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var matcher, err = requests.GetMatcher(r)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if matcher != nil {
		var matched = 0
		for i := 0; i < len(collection); i++ {
			if matcher(collection[i]) {
				collection[matched] = collection[i]
				matched++
			}
		}
		collection = collection[0:matched]
	}

	_, brief := r.URL.Query()["brief"]
	var response = ToJSon(collection)
	if brief {
		var briefs = make([]string, len(collection), len(collection))
		for i := 0; i < len(collection); i++ {
			briefs[i] = collection[i].GetSelf()
		}
		response = ToJSon(briefs)
	} else {
		response = ToJSon(collection)
	}

	if statusCode := requests.CheckEtag(r, response.Etag); statusCode != 0 {
		w.WriteHeader(statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", response.Etag)
	_, _ = w.Write(response.Data)
}

func ServeResource(w http.ResponseWriter, r *http.Request, res Resource) {
	var jsonResponse = ToJSon(res)

	if statusCode := requests.CheckEtag(r, jsonResponse.Etag); statusCode != 0 {
		w.WriteHeader(statusCode)
		return
	}

	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", jsonResponse.Etag)
		_, _ = w.Write(jsonResponse.Data)
	} else if r.Method == "POST" {
		res.POST(w, r)
	} else if r.Method == "PATCH" {
		res.PATCH(w, r)
	} else if r.Method == "DELETE" {
		res.DELETE(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
