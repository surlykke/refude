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
)

type ETagHandler interface {
	ETag() string
}

type GETHandler interface {
	GET(w http.ResponseWriter, r *http.Request)
}

type PATCHHandler interface {
	PATCH(w http.ResponseWriter, r *http.Request)
}

type POSTHandler interface {
	POST(w http.ResponseWriter, r *http.Request)
}

type DELETEHandler interface {
	DELETE(w http.ResponseWriter, r *http.Request)
}


func ServeHTTP(res interface{},  w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if getHandler, ok := res.(GETHandler); ok {
			getHandler.GET(w, r)
			return
		}
	} else if r.Method == "PATCH" {
		if patchHandler, ok := res.(PATCHHandler); ok {
			patchHandler.PATCH(w, r)
			return
		}
	} else if r.Method == "POST" {
		if postHandler, ok := res.(POSTHandler); ok {
			postHandler.POST(w, r)
			return
		}
	} else if r.Method == "DELETE" {
		if deleteHandler, ok := res.(DELETEHandler); ok {
			deleteHandler.DELETE(w, r)
			return
		}
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

func JsonGET(res interface{}, w http.ResponseWriter) {
	if bytes, err := json.Marshal(res); err != nil {
		panic("Could not json-marshal")
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

func GetSingleQueryParameter(r *http.Request, parameterName string, fallbackValue string) string {
	if len(r.URL.Query()[parameterName]) == 0 {
		return fallbackValue
	} else {
		return r.URL.Query()[parameterName][0]
	}
}
