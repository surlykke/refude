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

type Resource interface {
	GetSelf() StandardizedPath
	GetMt() MediaType
	POST(w http.ResponseWriter, r *http.Request)
	PATCH(w http.ResponseWriter, r *http.Request)
	DELETE(w http.ResponseWriter, r *http.Request)
}

type Executer func()

type ResourceAction struct {
	Description string
	IconName    string
	Executer    Executer `json:"-"`
}

type Relation string

const (
	Self               Relation = "self"
	Related                     = "related"
	Associated                  = "http://relations.refude.org/associated"
	DefaultApplication          = "http://relations.refude.org/default_application"
	SNI_MENU                    = "http://relations.refude.org/sni_menu"
)

type AbstractResource struct {
	Self            StandardizedPath          `json:"_self"` // Convenience - is also contained in Links
	Links           []Link                    `json:"_links"`
	Mt              MediaType                 `json:"-"`
	ResourceActions map[string]ResourceAction `json:"_actions,omitempty"`
}

type Link struct {
	Href  StandardizedPath `json:"href"`
	Rel   Relation         `json:"rel"` // We never have more than one relation on a link - we'll make a new link with same href
	Type  MediaType        `json:",omitempty"`
	Title string           `json:",omitempty"`
}

func MakeAbstractResource(SelfLink StandardizedPath, mt MediaType) AbstractResource {
	return AbstractResource{
		Self:            SelfLink,
		Links:           []Link{{Href: SelfLink, Rel: Self}},
		Mt:              mt,
		ResourceActions: make(map[string]ResourceAction),
	}
}

func (ar *AbstractResource) GetSelf() StandardizedPath {
	for _, link := range ar.Links {
		if link.Rel == Self {
			return link.Href
		}
	}

	panic("Resource has no self link")
}

func (ar *AbstractResource) GetMt() MediaType {
	return ar.Mt
}

func (r *AbstractResource) LinkTo(target StandardizedPath, relation Relation) {
	r.Links = append(r.Links, Link{Href: target, Rel: relation})
}

func (ar *AbstractResource) POST(w http.ResponseWriter, r *http.Request) {
	var actionId = requests.GetSingleQueryParameter(r, "action", "default")
	if action, ok := ar.ResourceActions[actionId]; ok {
		action.Executer()
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}

func (ar *AbstractResource) PATCH(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (ar *AbstractResource) DELETE(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

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
