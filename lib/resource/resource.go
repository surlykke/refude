// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
)

type Action struct {
	Id    string
	Title string
	Icon  string
}

type Resource struct {
	Self       link.Href   `json:"self"`
	Post       link.List   `json:"post"`
	Delete     *link.Link  `json:"delete,omitempty"`
	Patch      link.Href   `json:"patch,omitempty"`
	Links      link.Href   `json:"links,omitempty"`
	Searchable bool        `json:"searchable"`
	Path       string      `json:"-"`
	Title      string      `json:"title"`
	Comment    string      `json:"comment,omitempty"`
	Icon       link.Href   `json:"icon,omitempty"`
	Profile    string      `json:"profile"`
	Data       interface{} `json:"data"`
	Keywords   []string    `json:",omitempty"`
}

func MakeResource(path, title, comment, iconName, profile string, data interface{}) *Resource {
	var res = &Resource{
		Self:    link.Href(path),
		Post:    link.List{},
		Path:    path,
		Title:   title,
		Comment: comment,
		Icon:    link.IconUrl(iconName),
		Profile: profile,
		Data:    data,
	}

	if postable, ok := data.(Postable); ok {
		for _, act := range postable.GetPostActions() {
			res.Post = append(res.Post, link.Link{
				Href:  link.Href(path + "?action=" + act.Id),
				Title: act.Title,
				Icon:  link.IconUrl(act.Icon),
			})
		}
	}

	if deleteable, ok := data.(Deleteable); ok {
		var act = deleteable.GetDeleteAction()
		res.Delete = &link.Link{
			Href:  link.Href(path),
			Title: act.Title,
			Icon:  link.IconUrl(act.Icon),
		}
	}

	if linkable, ok := data.(Linkable); ok {
		res.Links = link.Href(path + "?want=links")
		res.Searchable = linkable.IsSearchable()
	}

	return res
}

func (res *Resource) MakeRankedLink(rank int) link.Link {
	return link.MakeRanked2(res.Self, res.Title, res.Icon, res.Profile, rank)
}

type Postable interface {
	GetPostActions() []Action
	DoPost(w http.ResponseWriter, r *http.Request)
}

type Deleteable interface {
	GetDeleteAction() *Action
	DoDelete(w http.ResponseWriter, r *http.Request)
}

type Linkable interface {
	IsSearchable() bool
	GetLinks(term string) link.List
}

func (res *Resource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if shouldGetLinks(r, res.Data) {
			var term = requests.GetSingleQueryParameter(r, "search", "")
			respond.AsJson(w, res.Data.(Linkable).GetLinks(term))
		} else {
			respond.AsJson(w, res)
		}
	} else if shouldDoPost(r.Method, res.Data) {
		res.Data.(Postable).DoPost(w, r)
	} else if shouldDoDelete(r.Method, res.Data) {
		res.Data.(Deleteable).DoDelete(w, r)
	} else {
		respond.NotAllowed(w)
	}
}

func shouldDoPost(method string, data interface{}) bool {
	if method == "POST" {
		if _, ok := data.(Postable); ok {
			return true
		}
	}
	return false
}

func shouldDoDelete(method string, data interface{}) bool {
	if method == "DELETE" {
		if _, ok := data.(Deleteable); ok {
			return true
		}
	}
	return false
}

func shouldGetLinks(r *http.Request, data interface{}) bool {
	if requests.GetSingleQueryParameter(r, "want", "") == "links" {
		if _, ok := data.(Linkable); ok {
			return true
		}
	}
	return false
}

type dataSlice []*Resource
