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

	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/respond"
)

type Resource interface {
	GetPath() string
	GetTitle() string
	GetComment() string
	GetIconUrl() string
	GetSearchUrl() string
	GetProfile() string
	GetLinks() LinkList
	GetKeywords() []string
	OmitFromSearch() bool
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
	Title    string   `json:"title"`
	Comment  string   `json:"comment,omitempty"`
	Profile  string   `json:"profile"`
	Links    LinkList `json:"links"`
	Keywords []string `json:"-"`
}

func MakeBase(path, title, comment, iconUrl, profile string) *ResourceData {
	var br = ResourceData{
		Path:    path,
		Title:   title,
		Comment: comment,
		Profile: profile,
		Links:   LinkList{{Href: "http://localhost:7938" + path, Relation: relation.Self}},
	}
	if iconUrl != "" {
		br.SetIconHref(iconUrl)
	}
	return &br
}

func (this *ResourceData) GetPath() string       { return this.Path }
func (this *ResourceData) GetTitle() string      { return this.Title }
func (this *ResourceData) GetComment() string    { return this.Comment }
func (this *ResourceData) GetIconUrl() string    { return this.Links.get(relation.Icon).Href }
func (this *ResourceData) GetSearchUrl() string  { return this.Links.get(relation.Search).Href }
func (this *ResourceData) GetProfile() string    { return this.Profile }
func (this *ResourceData) GetLinks() LinkList    { return this.Links }
func (this *ResourceData) GetKeywords() []string { return this.Keywords }
func (this *ResourceData) OmitFromSearch() bool  { return false }

func (this *ResourceData) SetIconHref(iconUrl string) {
	this.Links = this.Links.set(iconUrl, "", "", relation.Icon)
}

func (this *ResourceData) SetSearchHref(searchUrl string) {
	this.Links = this.Links.set(searchUrl, "", "", relation.Search)
}

func (this *ResourceData) AddLink(href string, title string, iconUrl string, rel relation.Relation) {
	this.Links = this.Links.add(href, title, iconUrl, rel)
}

var httpLocalHost7838 = []byte("http://localhost:7938")

// --------------------- Link --------------------------------------

type Link struct {
	Href     string            `json:"href"`
	Title    string            `json:"title,omitempty"`
	IconUrl  string            `json:"icon,omitempty"`
	Relation relation.Relation `json:"rel,omitempty"`
}

type LinkList []Link

func (ll LinkList) add(href, title, iconUrl string, relation relation.Relation) LinkList {
	var res = slices.Clone(ll)
	return append(res, Link{Href: normalizeHref(href), Title: title, IconUrl: iconUrl, Relation: relation})
}

// Ensure only one link with given relation in list
func (ll LinkList) set(href, title, iconUrl string, relation relation.Relation) LinkList {
	var res = make(LinkList, len(ll)+1, len(ll)+1)
	var pos = 0
	for i := 0; i < len(ll); i++ {
		if ll[i].Relation != relation {
			res[pos] = ll[i]
			pos++
		}
	}
	res[pos] = Link{Href: normalizeHref(href), Title: title, IconUrl: iconUrl, Relation: relation}
	res = res[0 : pos+1]
	return res
}

// Gets first found link with given relation
func (ll LinkList) get(relation relation.Relation) Link {
	for _, l := range ll {
		if l.Relation == relation {
			return l
		}
	}
	return Link{}
}

func normalizeHref(href string) string {
	if strings.HasPrefix(href, "/") {
		return "http://localhost:7938" + href
	} else {
		return href
	}
}

// -------------- Serve -------------------------

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
