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

	"github.com/sahilm/fuzzy"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/respond"
)

type Resource interface {
	Data() *ResourceData
	OmitFromSearch() bool
}

type ResourceData struct {
	Path     string
	Title    string              `json:"title"`
	Comment  string              `json:"comment,omitempty"`
	Type     mediatype.MediaType `json:"type"`
	Links    LinkList            `json:"links"`
	Keywords []string            `json:"keywords"`
}

func MakeBase(path, title, comment, iconUrl string, mType mediatype.MediaType) *ResourceData {
	var br = ResourceData{
		Path:    path,
		Title:   title,
		Comment: comment,
		Type:    mType,
		Links:   LinkList{{Href: "http://localhost:7938" + path, Relation: relation.Self}},
	}
	if iconUrl != "" {
		br.SetIconHref(iconUrl)
	}
	return &br
}

func (this *ResourceData) Data() *ResourceData {
	return this
}

func (this *ResourceData) GetLink(rel relation.Relation) Link {
	for _, link := range this.Links {
		if rel == link.Relation {
			return link
		}
	}
	return Link{}
}

func (this *ResourceData) GetLinks(relations ...relation.Relation) []Link {
	if len(relations) == 0 {
		return slices.Clone(this.Links)
	} else {
		var result = make([]Link, 0, len(this.Links))
		for _, lnk := range this.Links {
			for _, rel := range relations {
				if lnk.Relation == rel {
					result = append(result, lnk)
				}
			}
		}
		return result
	}
}

func (this *ResourceData) OmitFromSearch() bool {
	return false
}

func (this *ResourceData) SetIconHref(iconUrl string) {
	for i := 0; i < len(this.Links); i++ {
		if this.Links[i].Relation == relation.Icon {
			this.Links[i].Href = iconUrl
			return
		}
	}
	this.Links = append(this.Links, Link{Href: iconUrl, Relation: relation.Icon})
}

func (this *ResourceData) AddLink(href string, title string, iconUrl string, rel relation.Relation) {
	this.Links = append(this.Links, Link{Href: href, Title: title, IconUrl: iconUrl, Relation: rel})
}

var httpLocalHost7838 = []byte("http://localhost:7938")

// --------------------- Link --------------------------------------

type Link struct {
	Href     string              `json:"href"`
	Title    string              `json:"title,omitempty"`
	IconUrl  string              `json:"icon,omitempty"`
	Relation relation.Relation   `json:"rel,omitempty"`
	Type     mediatype.MediaType `json:"type,omitempty"`
}

type LinkList []Link

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
		var matches = fuzzy.FindFrom(term, ll)
		var sorted = make(LinkList, len(matches), len(matches))
		for i, match := range matches {
			sorted[i] = ll[match.Index]
		}
		return sorted
	}
}

func LinkTo(res Resource) Link {
	var lnk = res.Data().GetLink(relation.Self)
	lnk.Relation = relation.Related
	lnk.Title = res.Data().Title
	lnk.IconUrl = res.Data().GetLink(relation.Icon).Href
	lnk.Type = res.Data().Type
	return lnk
}

func NormalizeHref(href string) string {
	if strings.HasPrefix(href, "/") {
		return "http://localhost:7938" + href
	} else {
		return href
	}
}

func GetPath(l Link) string {
	if strings.HasPrefix(l.Href, "http://localhost:7938") {
		return l.Href[len("http://localhost:7938"):]
	} else if strings.HasPrefix(l.Href, "/") {
		return l.Href
	} else {
		return ""
	}
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
