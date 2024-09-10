// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package resource

import (
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strings"

	"github.com/sahilm/fuzzy"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/respond"
)

type Resource interface {
	GetPath() string
	GetTitle() string
	GetComment() string
	GetProfile() string
	GetLinks() LinkList
	GetKeywords() []string
	Search(term string) LinkList
	OmitFromSearch() bool
}

type ResourceData struct {
	Path     string
	Title    string   `json:"title"`
	Comment  string   `json:"comment,omitempty"`
	Profile  string   `json:"profile"`
	Links    LinkList `json:"links"`
	Keywords []string `json:"keywords"`
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

func (this *ResourceData) GetPath() string {
	return this.Path
}

func (this *ResourceData) GetTitle() string {
	return this.Title
}

func (this *ResourceData) GetComment() string {
	return this.Comment
}

func (this *ResourceData) GetProfile() string {
	return this.Profile
}

func (this *ResourceData) GetLinks() LinkList {
	return this.Links
}

func (this *ResourceData) GetKeywords() []string {
	return this.Keywords
}

func (this *ResourceData) Search(term string) LinkList {
	var result = make(LinkList, 0, 10)
	for _, lnk := range this.GetLinks() {
		if lnk.Relation == relation.DefaultAction || lnk.Relation == relation.Action || lnk.Relation == relation.Delete {
			result = append(result, lnk)
		}
	}
	return result
}

func (this *ResourceData) OmitFromSearch() bool {
	return false
}

func (this *ResourceData) SetIconHref(iconUrl string) {
	this.Links = this.Links.set(iconUrl, "", "", relation.Icon)
}

func (this *ResourceData) SetSearchHref(searchUrl string) {
	this.Links = this.Links.set(searchUrl, "", "", relation.Search)
}

func (this *ResourceData) AddLink(href string, title string, iconUrl string, rel relation.Relation) {
	this.Links = this.Links.add(href, title, iconUrl, rel)
}

func (this *ResourceData) GetDefaultAction() (Link, bool) {
	for _, lnk := range this.Links {
		if lnk.Relation == relation.DefaultAction {
			return lnk, true
		}
	}
	return Link{}, false
}

var httpLocalHost7838 = []byte("http://localhost:7938")

// --------------------- Link --------------------------------------

type Link struct {
	Href     string            `json:"href"`
	Title    string            `json:"title,omitempty"`
	IconUrl  string            `json:"icon,omitempty"`
	Relation relation.Relation `json:"rel,omitempty"`
	Type     string            `json:"type,omitempty"`
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
func (ll LinkList) Get(relation relation.Relation) Link {
	for _, l := range ll {
		if l.Relation == relation {
			return l
		}
	}
	return Link{}
}

// Implement fuzzy.Source

func (ll LinkList) String(i int) string {
	return ll[i].Title
}

func (ll LinkList) Len() int {
	return len(ll)
}

func (ll LinkList) FilterAndSort(term string) LinkList {
	if term == "" {
		return ll
	} else {
		if lastSlash := strings.LastIndex(term, "/"); lastSlash > -1 {
			term = term[lastSlash+1:]
		}
		fmt.Println("Matching with", term)
		var matches = fuzzy.FindFrom(term, ll)
		var sorted = make(LinkList, len(matches), len(matches))
		for i, match := range matches {
			sorted[i] = ll[match.Index]
		}
		return sorted
	}
}

func LinkTo(res Resource) Link {
	var lnk = res.GetLinks().Get(relation.Self)
	lnk.Relation = relation.Related
	lnk.Title = res.GetTitle()
	lnk.IconUrl = res.GetLinks().Get(relation.Icon).Href
	return lnk
}

func normalizeHref(href string) string {
	if strings.HasPrefix(href, "/") {
		return "http://localhost:7938" + href
	} else {
		return href
	}
}

type RankedLink struct {
	Link
	Rank int
}

type RankedLinkList []RankedLink

func (this RankedLinkList) Sort() {
	sort.SliceStable(this, func(i, j int) bool { return this[i].Rank < this[j].Rank })
}

func (this RankedLinkList) GetLinksSorted() LinkList {
	this.Sort()
	var ll = make(LinkList, 0, len(this))
	for _, rl := range this {
		ll = append(ll, rl.Link)
	}
	return ll
}

// -------------- Serve -------------------------

type Postable interface {
	DoPost(w http.ResponseWriter, r *http.Request)
}

type Deleteable interface {
	DoDelete(w http.ResponseWriter, r *http.Request)
}

type Searchable interface {
	Search(term string) RankedLinkList
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
