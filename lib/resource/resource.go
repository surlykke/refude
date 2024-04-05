// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package resource

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/respond"
)

type Resource interface {
	Base() *BaseResource
	Search(searchTerm string) []Resource
	RelevantForSearch(term string) bool
}

type BaseResource struct {
	Path     string
	Title    string `json:"title"`
	Comment  string `json:"comment,omitempty"`
	IconUrl  string `json:"icon,omitempty"`
	Profile  string `json:"profile"`
	Links    []link.Link `json:"links"`
	Keywords []string `json:"-"`
}
/*	Links   []link.Link `json:"links"`
	Title   string      `json:"title"`
	Comment string      `json:"comment,omitempty"`
	Icon    string      `json:"icon,omitempty"`
	Profile string      `json:"profile"`
	Data    interface{} `json:"data"`
*/


func MakeBase(path, title, comment, iconUrl, profile string) *BaseResource {
	var br = BaseResource{
		Path:    path,
		Title:   title,
		Comment: comment,
		IconUrl: iconUrl,
		Profile: profile,
		Links:   []link.Link{{Href:"http://localhost:7938" + path, Relation:relation.Self}},
	}
	br.AddLink("", "", "", relation.Self)
	return &br
}

func (this *BaseResource) AddLink(href, title, iconUrl string, relation relation.Relation) {
	if href == "" { 
		href = this.Path
	} else if strings.HasPrefix(href, "?") {
		href = this.Path + href 
	}
	if strings.HasPrefix(href, "/") {
		href = "http://localhost:7938" + href
	}
	this.Links = append(this.Links, link.Link{Href: href, Title: title, IconUrl: iconUrl, Relation: relation}) 
}


func (this *BaseResource) Base() *BaseResource {
	return this
}

func (this *BaseResource) Search(searchTerm string) []Resource {
	return []Resource{}
}

func (br *BaseResource) RelevantForSearch(term string) bool {
	return false
}

func (br *BaseResource) ActionLinks() []link.Link {
	var filtered = make([]link.Link, 0, len(br.Links))
	for _, lnk := range br.Links {
		if lnk.Relation == relation.Action || lnk.Relation == relation.Delete {
			filtered = append(filtered, lnk)
		}
	}
	return filtered
}

func (br *BaseResource) Searchable() bool {
	for _, lnk := range br.Links {
		if lnk.Relation == relation.Search {
			return true
		}
	}
	return false
}

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

