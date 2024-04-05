// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package resource

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type Resource interface {
	Base() *BaseResource
	Search(searchTerm string) []Resource
	RelevantForSearch(term string) bool
}

type BaseResource struct {
	Path     string
	Title    string `json:"-"`
	Comment  string `json:"-"`
	IconUrl  string `json:"-"`
	Profile  string `json:"-"`
	Links    []link.Link
	Keywords []string `json:"profile,omitempty"`
}

func MakeBase(path, title, comment, iconUrl, profile string, search bool) BaseResource {
	var br = BaseResource{
		Path:    path,
		Title:   title,
		Comment: comment,
		Profile: profile,
		Links:   make([]link.Link, 0, 5),
	}

	br.SetIconUrl(iconUrl)
	br.AddLink(link.Link{Href: path, Relation: relation.Self})
	if search {
		br.AddLink(link.Link{Href: "/search?from=" + url.QueryEscape(path), Relation: relation.Search})
	}

	return br
}

func (this *BaseResource) SetIconUrl(iconUrl string) {
	if iconUrl != "" {
		if !(strings.HasPrefix(iconUrl, "http://") || strings.HasPrefix(iconUrl, "https://")) {
			if strings.Index(iconUrl, "/") > -1 {
				// So its a path..
				if strings.HasPrefix(iconUrl, "file:///") {
					iconUrl = iconUrl[7:]
				} else if strings.HasPrefix(iconUrl, "file://") {
					iconUrl = xdg.Home + "/" + iconUrl[7:]
				} else if !strings.HasPrefix(iconUrl, "/") {
					iconUrl = xdg.Home + "/" + iconUrl
				}
			}
			iconUrl = "http://localhost:7938/icon?name=" + url.QueryEscape(iconUrl)
		}
		this.IconUrl = iconUrl
	}
}

func (this *BaseResource) AddLink(lnk link.Link) {
	this.Links = append(this.Links, lnk)
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
		respond.AsJson(w, BuildJsonRepresentation(res))
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
		var jsonReps = make([]jsonRepresentation, 0, len(list))
		for _, res := range list {
			jsonReps = append(jsonReps, BuildJsonRepresentation(res))
		}
		respond.AsJson(w, jsonReps)
	} else {
		respond.NotAllowed(w)
	}
}

type jsonRepresentation struct {
	Links   []link.Link `json:"links"`
	Title   string      `json:"title"`
	Comment string      `json:"comment,omitempty"`
	Icon    string      `json:"icon,omitempty"`
	Profile string      `json:"profile"`
	Data    interface{} `json:"data"`
}

func BuildJsonRepresentation(res Resource) jsonRepresentation {
	var wrapper = jsonRepresentation{}
	wrapper.Links = res.Base().Links
	wrapper.Data = res
	wrapper.Title = res.Base().Title
	wrapper.Comment = res.Base().Comment
	wrapper.Icon = string(res.Base().IconUrl)
	wrapper.Profile = res.Base().Profile
	return wrapper
}
