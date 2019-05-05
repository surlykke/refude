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
	ServeHttp(w http.ResponseWriter, r *http.Request)
}

type ResourceCollection interface {
	Get(path string) Resource
}

func ToBytesAndEtag(res interface{}) ([]byte, string) {
	var bytes []byte
	var etag string
	var err error
	if bytes, err = json.Marshal(res); err != nil {
		panic(fmt.Sprintln(err))
	}
	etag = fmt.Sprintf("\"%x\"", sha1.Sum(bytes))

	return bytes, etag
}

/*func ServeCollection(w http.ResponseWriter, r *http.Request, collection []Resource) {
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
}*/
