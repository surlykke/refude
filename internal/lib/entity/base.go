// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package entity

import (
	"strings"

	"github.com/surlykke/refude/internal/lib/translate"
)

type Servable interface {
	GetBase() *Base
	OmitFromSearch() bool
}

type Base struct {
	Title      string `json:"title"`
	Summary    string `json:"summary"`
	Icon       string `json:"icon"`
	Keywords   []string
	Action     *Action  `json:"-"`
	Delete     *Action  `json:"-"`
	AltActions []Action `json:"-"`
	Links      map[Relation][]Link
	Path       string `json:"path"`
}

type Action struct {
	Id   string
	Name string
	Icon string
	Url  string
}

func MakeBase(title string, summary string, icon string, keywords ...string) *Base {
	icon = adjustIcon(icon)
	return &Base{
		Title:    translate.Text(title),
		Summary:  translate.Text(summary),
		Icon:     icon,
		Keywords: translate.Texts(keywords),
		Links:    map[Relation][]Link{Self: {Link{}}},
	}
}

func adjustIcon(icon string) string {
	icon = strings.TrimSpace(icon)
	if icon == "" {
		return ""
	} else if strings.HasPrefix(icon, "http://") || strings.HasPrefix(icon, "https://") || strings.HasPrefix(icon, "/icon?name=") {
		return icon
	} else {
		return "/icon?name=" + icon
	}
}

func (this *Base) OmitFromSearch() bool {
	return false
}

func (this *Base) GetBase() *Base {
	return this
}


func (this *Base) AddAction(title string, icon string, relation Relation, query string) {
	this.Links[relation] = append(this.Links[relation], Link{Title: title, Icon: icon, query: query})
}

// Call before serving
func (this *Base) SetPath(path string) {
	this.Path = path // TODO: Necessary ?
	for _, linklist := range this.Links {
		for i := 0; i < len(linklist); i++ {
			linklist[i].Href = path
			if "" == linklist[i].query {
				linklist[i].Href = path
			} else {
				linklist[i].Href = path + "?" + linklist[i].query
			}
		}
	}
}

func (this *Base) FirstAction() *Link {
	if len(this.Links[OrgRefudeAction]) > 0 {
		return &this.Links[OrgRefudeAction][0]
	} else {
		return nil
	}
}

func (this *Base) SecondaryActions() []Link {
	if len(this.Links[OrgRefudeAction]) > 1 {
		return this.Links[OrgRefudeAction][1:]
	} else {
		return nil
	}
}	

func (this *Base) DeleteAction() *Link {
	if len(this.Links[OrgRefudeDelete]) > 0 {
		return &this.Links[OrgRefudeDelete][0]
	} else {
		return nil
	}
}



type Link struct {
	Href     string   `json:"href"`
	Title    string   `json:"title,omitempty"`
	Icon     string   `json:"icon,omitempty"`
	query    string   
}

type Relation string

const (
	Self            = "self"
	Icon            = "icon"
	Related         = "related"
	OrgRefudeAction = "org.refude.action"
	OrgRefudeDelete = "org.refude.delete"
	OrgRefudeMenu   = "org.refude.menu"
)

// -------------- Serve -------------------------

type Postable interface {
	DoPost(string) (bool, error)
}

type Deleteable interface {
	DoDelete() error
}
