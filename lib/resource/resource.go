// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package resource

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

type Resource interface {
	GetId() string
	Presentation() (title string, comment string, iconName string, profile string)
	Actions() link.ActionList
	DeleteAction() (title string, ok bool)
	Links(searchTerm string) link.List
	RelevantForSearch() bool
	GetKeywords() []string
}

type BaseResource struct {
	Id       string
	Title    string
	Comment  string `json:",omitempty"`
	IconName string `json:",omitempty"`
	Profile  string
	Keywords []string
}

func (br *BaseResource) GetId() string {
	return br.Id
}

func (br *BaseResource) Presentation() (title string, comment string, iconName string, profile string) {
	return br.Title, br.Comment, br.IconName, br.Profile
}

func (br *BaseResource) Actions() link.ActionList {
	return link.ActionList{}
}

func (br *BaseResource) DeleteAction() (string, bool) {
	return "", false
}

func (br *BaseResource) Links(searchTerm string) link.List {
	return link.List{}
}

func (br *BaseResource) RelevantForSearch() bool {
	return true
}

func (br *BaseResource) GetKeywords() []string {
	return br.Keywords
}

func LinkTo(res Resource, context string, rank int) link.Link {
	var path = fmt.Sprint(context, res.GetId())
	var title, _, iconName, profile = res.Presentation()
	return link.MakeRanked(path, title, iconName, profile, rank)
}

type Postable interface {
	DoPost(w http.ResponseWriter, r *http.Request)
}

type Deleteable interface {
	DoDelete(w http.ResponseWriter, r *http.Request)
}

type ResourceRepo interface {
	GetResources() []Resource
	GetResource(path string) Resource
	Search(term string, threshold int) link.List
}


func SingleResourceServer(res Resource, context string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ServeSingleResource(w, r, res, context)
	}
}

func ServeSingleResource(w http.ResponseWriter, r *http.Request, res Resource, context string) {
	if r.Method == "GET" {
		var linkSearchTerm = requests.GetSingleQueryParameter(r, "search", "")
		respond.AsJson(w, buildJsonRepresentation(res, context, linkSearchTerm))
	} else if postable, ok := res.(Postable); ok && r.Method == "POST" {
		postable.DoPost(w, r)
	} else if deletable, ok := res.(Deleteable); ok && r.Method == "DELETE" {
		deletable.DoDelete(w, r)
	} else {
		respond.NotAllowed(w)
	}
}


type jsonRepresentation struct {
	Self    link.Href   `json:"self"`
	Links   link.List   `json:"links"`
	Title   string      `json:"title"`
	Comment string      `json:"comment,omitempty"`
	Icon    link.Href   `json:"icon,omitempty"`
	Profile string      `json:"profile"`
	Data    interface{} `json:"data"`
}

func buildJsonRepresentation(res Resource, context, searchTerm string) jsonRepresentation {
	var wrapper = jsonRepresentation{}
	wrapper.Self = link.Href(context + res.GetId())
	wrapper.Links = buildFilterAndRewriteLinks(res, context, searchTerm)
	wrapper.Data = res
	var iconName string
	wrapper.Title, wrapper.Comment, iconName, wrapper.Profile = res.Presentation()
	wrapper.Icon = link.IconUrl(iconName)
	return wrapper
}

func buildFilterAndRewriteLinks(res Resource, context, searchTerm string) link.List {
	var list = make(link.List, 0, 10)
	for _, action := range res.Actions() {
		var href = context + res.GetId()
		if action.Name != "" {
			href += "?action=" + action.Name
		}
		if searchutils.Match(searchTerm, action.Name) < 0 {
			continue
		}
		list = append(list, link.Make(href, action.Title, action.IconName, relation.Action))
	}
	if deleteTitle, ok := res.DeleteAction(); ok {
		if searchutils.Match(searchTerm, deleteTitle) > -1 {
			list = append(list, link.Make(context+res.GetId(), deleteTitle, "", relation.Delete))
		}
	}

	var lnks link.List = res.Links(searchTerm)

	for _, lnk := range lnks {
		if !strings.HasPrefix(string(lnk.Href), "/") {
			lnk.Href = link.Href(context) + lnk.Href
		}
		list = append(list, lnk)
	}

	return list
}



