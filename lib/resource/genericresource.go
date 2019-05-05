// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"net/http"
	"sync"

	"github.com/surlykke/RefudeServices/lib/requests"
)

type Executer func()

type ResourceAction struct {
	Description string
	IconName    string
	Executer    Executer `json:"-"`
}

type JsonResourceData interface {
	POST(w http.ResponseWriter, r *http.Request)
	PATCH(w http.ResponseWriter, r *http.Request)
	DELETE(w http.ResponseWriter, r *http.Request)
}

// For embedding
type GeneralTraits struct {
	Self       string `json:"_self,omitempty"`
	RefudeType string `json:"_refudetype,omitempty"`
}

func (gt *GeneralTraits) GetSelf() string {
	return gt.Self
}

type DefaultMethods struct {
	Actions map[string]ResourceAction `json:"_actions,omitempty"`
}

func (dm *DefaultMethods) AddAction(actionId string, action ResourceAction) {
	if dm.Actions == nil {
		dm.Actions = make(map[string]ResourceAction)
	}
	dm.Actions[actionId] = action
}

func (dm *DefaultMethods) POST(w http.ResponseWriter, r *http.Request) {
	if dm.Actions == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		var actionId = requests.GetSingleQueryParameter(r, "action", "default")
		if action, ok := dm.Actions[actionId]; ok {
			action.Executer()
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
	}
}

func (dm *DefaultMethods) PATCH(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (dm *DefaultMethods) DELETE(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

type JsonCache struct {
	sync.Mutex
	json []byte
	etag string
}

type JsonResource struct {
	Data  JsonResourceData
	cache *JsonCache
}

func MakeJsonResource(data JsonResourceData) JsonResource {
	return JsonResource{Data: data, cache: &JsonCache{}}
}

func (jr JsonResource) getJsonData() ([]byte, string) {
	jr.cache.Lock()
	defer jr.cache.Unlock()
	if jr.cache.json == nil {
		jr.cache.json, jr.cache.etag = ToBytesAndEtag(jr.Data)
	}
	return jr.cache.json, jr.cache.etag
}

func (jr JsonResource) ServeHttp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var bytes, etag = jr.getJsonData()
		if statusCode := requests.CheckEtag(r, etag); statusCode != 0 {
			w.WriteHeader(statusCode)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("ETag", etag)
			_, _ = w.Write(bytes)
		}
	case "POST":
		jr.Data.POST(w, r)
	case "PATCH":
		jr.Data.PATCH(w, r)
	case "DELETE":
		jr.Data.DELETE(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
