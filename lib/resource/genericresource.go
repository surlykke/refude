// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/serialize"
)

type Executer func()

type ResourceAction struct {
	Description string
	IconName    string
	Executer    Executer `json:"-"`
}

type GenericResource struct {
	Links           []Link                    `json:"_links"`
	Mt              MediaType                 `json:"-"`
	ResourceActions map[string]ResourceAction `json:"_actions,omitempty"`
	etag            string
}

func MakeGenericResource(SelfLink string, mt MediaType) GenericResource {
	return GenericResource{
		Links:           []Link{{Href: SelfLink, Rel: Self}},
		Mt:              mt,
		ResourceActions: make(map[string]ResourceAction),
	}
}

func (gr *GenericResource) GetSelf() string {
	for _, link := range gr.Links {
		if link.Rel == Self {
			return link.Href
		}
	}

	panic("Resource has no self link")
}

func (gr *GenericResource) GetMt() MediaType {
	return gr.Mt
}

func (gr *GenericResource) GetEtag() string {
	return gr.etag
}

func (gr *GenericResource) SetEtag(etag string) {
	gr.etag = etag
}

func (gr *GenericResource) LinkTo(target string, relation Relation) {
	gr.Links = append(gr.Links, Link{Href: target, Rel: relation})
}

func (gr *GenericResource) POST(w http.ResponseWriter, r *http.Request) {
	var actionId = requests.GetSingleQueryParameter(r, "action", "default")
	if action, ok := gr.ResourceActions[actionId]; ok {
		action.Executer()
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}

func (gr *GenericResource) PATCH(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (gr *GenericResource) DELETE(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func CalculateEtag(bp ByteProducer) string {
	var sha1Hasher = sha1.New()
	bp.WriteBytes(sha1Hasher)
	return fmt.Sprintf("%X", sha1Hasher.Sum(nil))
}

type ByteProducer interface {
	WriteBytes(io.Writer)
}

func (gr *GenericResource) WriteBytes(w io.Writer) {
	for _, link := range gr.Links {
		serialize.String(w, string(link.Href))
		serialize.String(w, string(link.Type))
		serialize.String(w, string(link.Title))
		serialize.String(w, string(link.Rel))
	}
	for id, action := range gr.ResourceActions {
		serialize.String(w, id)
		serialize.String(w, action.Description)
		serialize.String(w, action.IconName)
	}
}
