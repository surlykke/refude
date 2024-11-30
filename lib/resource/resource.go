// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package resource

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/href"
	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/path"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/respond"
)

type Resource interface {
	Data() *ResourceData
	OmitFromSearch() bool
}

type ResourceData struct {
	Path     path.Path `json:"path"`
	Links    []Link
	Keywords []string `json:"keywords"`
}

func MakeBase(path path.Path, title, comment string, icon icon.Name, mType mediatype.MediaType) *ResourceData {
	var br = ResourceData{
		Path:  path,
		Links: []Link{{Href: href.Of(path), Title: title, Comment: comment, Icon: icon, Type: mType, Relation: relation.Self}},
	}
	return &br
}

func (this *ResourceData) Data() *ResourceData {
	return this
}

func (this *ResourceData) Link() Link {
	if len(this.Links) < 1 || this.Links[0].Relation != relation.Self {
		panic("ResourceData does not hold self link as first link")
	}
	return this.Links[0]
}

// ------------ Don't call after published ------------------

func (this *ResourceData) AddAction(actionId string, title string, comment string, iconName icon.Name, keywords ...string) {
	var lnk = Link{Href: href.Of(this.Path).P("action", actionId), Title: title, Comment: comment, Icon: iconName, Relation: relation.Action}
	this.Links = append(this.Links, lnk)
}

func (this *ResourceData) AddDeleteAction(actionId string, title string, comment string, iconName icon.Name) {
	this.Links = append(this.Links, Link{Href: href.Of(this.Path).P("action", actionId), Title: title, Comment: comment, Icon: iconName, Relation: relation.Delete})
}

// ----------------------------------------------------------

func (this *ResourceData) OmitFromSearch() bool {
	return false
}

func (this *ResourceData) GetLinks(relations ...relation.Relation) []Link {
	var result = make([]Link, 0, len(this.Links))
	for _, l := range this.Links {
		for _, rel := range relations {
			if l.Relation == rel {
				result = append(result, l)
			}
		}
	}

	return result
}

// --------------------- Link --------------------------------------

type Link struct {
	Href     href.Href           `json:"href"`
	Title    string              `json:"title,omitempty"`
	Comment  string              `json:"comment,omitempty"`
	Icon     icon.Name           `json:"icon,omitempty"`
	Relation relation.Relation   `json:"rel,omitempty"`
	Type     mediatype.MediaType `json:"type,omitempty"`
}

// -------------- Serve -------------------------

type Postable interface {
	DoPost(w http.ResponseWriter, r *http.Request)
}

type Deleteable interface {
	DoDelete(w http.ResponseWriter, r *http.Request)
}

func ServeSingleResource(w http.ResponseWriter, r *http.Request, res Resource) {
	if r.Method == "GET" {
		respond.AsJson(w, res)
	} else if postable, ok := res.(Postable); ok && r.Method == "POST" {
		postable.DoPost(w, r)
	} else if deletable, ok := res.(Deleteable); ok && r.Method == "DELETE" {
		deletable.DoDelete(w, r)
	} else {
		respond.NotAllowed(w)
	}
}

func ServeList(w http.ResponseWriter, r *http.Request, list []Resource) {
	if r.Method == "GET" {
		respond.AsJson(w, list)
	} else {
		respond.NotAllowed(w)
	}
}
