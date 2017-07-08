// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"encoding/json"
	"net/http"
	"sync"
	"reflect"
)

type ResourceHandler func(this *Resource, w http.ResponseWriter, r *http.Request)

func defaultHandler(this *Resource, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type Resource struct {
	Data     interface{}
	ETag     string
	mutex    sync.Mutex
	GET    ResourceHandler
	PATCH  ResourceHandler
	POST   ResourceHandler
	DELETE ResourceHandler
}

func (res *Resource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handler ResourceHandler = nil
	switch r.Method {
	case "GET": handler = res.GET
	case "PATCH": handler = res.PATCH
	case "POST": handler = res.POST
	case "DELETE": handler = res.DELETE
	}

	if handler != nil {
		handler(res, w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (res *Resource) Equal(other *Resource) bool {
	return reflect.DeepEqual(res.Data, other.Data)
}

func JsonResource(data interface{}, postHandler ResourceHandler) *Resource {
	return & Resource{
		Data: data,
		GET: JsonGET,
		POST: postHandler,
	}
}

func JsonGET(this *Resource, w http.ResponseWriter, r *http.Request) {
	bytes, err := json.Marshal(this.Data)
	if err != nil {
		panic("Could not json-marshal")
	};
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func GetSingleQueryParameter(r *http.Request, parameterName string, fallbackValue string) string {
	if len(r.URL.Query()[parameterName]) == 0 {
		return fallbackValue
	} else {
		return r.URL.Query()[parameterName][0]
	}
}
