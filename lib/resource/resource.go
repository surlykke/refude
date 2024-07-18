// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package resource

import (
	"net/http"
	"slices"
	"strings"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

type Resource interface {
	GetPath() string
	GetTitle() string
	GetComment() string
	GetIconUrl() string
	GetProfile() string
	GetLinks() []link.Link
	GetActionLinks(string) []link.Link
	GetKeywords() []string

	OmitFromSearch() bool
}

type HasBase interface {
	ResourceData
}

type RankedResource struct {
	Rank int
	Res  Resource
}

type RRList []RankedResource

func (rrList RRList) GetResources() []Resource {
	slices.SortFunc(rrList, cmp)
	var resList = make([]Resource, 0, len(rrList))
	for _, rr := range rrList {
		resList = append(resList, rr.Res)
	}
	return resList
}

func cmp(r1, r2 RankedResource) int {
	if r1.Rank != r2.Rank {
		return r1.Rank - r2.Rank
	} else {
		return strings.Compare(r1.Res.GetPath(), r2.Res.GetPath())
	}
}

type ResourceData struct {
	Path     string
	Title    string      `json:"title"`
	Comment  string      `json:"comment,omitempty"`
	IconUrl  string      `json:"icon,omitempty"`
	Profile  string      `json:"profile"`
	Links    []link.Link `json:"links"`
	Keywords []string    `json:"-"`
}

func MakeBase(path, title, comment, iconUrl, profile string) *ResourceData {
	var br = ResourceData{
		Path:    path,
		Title:   title,
		Comment: comment,
		IconUrl: iconUrl,
		Profile: profile,
		Links:   []link.Link{{Href: "http://localhost:7938" + path, Relation: relation.Self}},
	}
	return &br
}

func (this *ResourceData) GetPath() string       { return this.Path }
func (this *ResourceData) GetTitle() string      { return this.Title }
func (this *ResourceData) GetComment() string    { return this.Comment }
func (this *ResourceData) GetIconUrl() string    { return this.IconUrl }
func (this *ResourceData) GetProfile() string    { return this.Profile }
func (this *ResourceData) GetLinks() []link.Link { return this.Links }
func (this *ResourceData) GetKeywords() []string { return this.Keywords }
func (this *ResourceData) OmitFromSearch() bool  { return false }

func (this *ResourceData) AddLink(href, title, iconUrl string, relation relation.Relation) {
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

// -----------------------------------------------------

func (br *ResourceData) GetActionLinks(searchTerm string) []link.Link {
	var filtered = make([]link.Link, 0, len(br.Links))
	for _, lnk := range br.Links {
		if (lnk.Relation == relation.Action || lnk.Relation == relation.Delete) && searchutils.Match(searchTerm, lnk.Title) >= 0 {
			filtered = append(filtered, lnk)
		}
	}
	return filtered
}

type Postable interface {
	DoPost(w http.ResponseWriter, r *http.Request)
}

type Deleteable interface {
	DoDelete(w http.ResponseWriter, r *http.Request)
}

type Searchable interface {
	Search(term string) []Resource
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
