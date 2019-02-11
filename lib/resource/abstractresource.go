// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"github.com/surlykke/RefudeServices/lib/requests"
	"net/http"
)


type ResourceAction struct {
	Description   string
	IconName  	  string
	Executer      Executer `json:"-"`
}


func (a *ResourceAction) POST(w http.ResponseWriter, r *http.Request) {
	if a.Executer != nil {
		a.Executer()
	}
}

type Relation string


const (
	Self Relation = "self"
	Related = "related"
	Associated = "http://relations.refude.org/associated"
	DefaultApplication = "http://relations.refude.org/default_application"
)

type AbstractResource struct {
	Self            StandardizedPath          `json:"_self,omitempty"`
	Links           []Link                    `json:"_links"`
	Mt              MediaType                 `json:"-"`
	ResourceActions map[string]ResourceAction `json:"_actions, omitempty"`
}

type Link struct {
	Href  StandardizedPath `json:"href"`
	Rel   Relation         `json:"rel"` // We never have more than one relation on a link - we'll make a new link with same href
	Type  MediaType        `json:",omitempty"`
	Title string           `json:",omitempty"`
}

func MakeAbstractResource(SelfLink StandardizedPath, mt MediaType) AbstractResource {
	return AbstractResource{
		Self:  SelfLink,
		Links: []Link{Link{Href: SelfLink, Rel: Self}},
		Mt:    mt,
		ResourceActions: make(map[string]ResourceAction),
	}
}

func (ar *AbstractResource) GetSelf() StandardizedPath {
	return ar.Self
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


