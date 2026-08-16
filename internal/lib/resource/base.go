// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package resource

import (
	"net/http"
	"strings"

	"github.com/surlykke/refude/internal/lib/respond"
	"github.com/surlykke/refude/internal/lib/translate"
)

type Servable interface {
	GetBase() Base
	OmitFromSearch() bool
}

type Base struct {
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Icon     string `json:"icon"`
	Keywords []string
	Links    map[Relation][]Link
	Path     string `json:"path"`
}

func MakeBase(path string, title string, summary string, icon string, keywords ...string) Base {
	icon = adjustIcon(icon)
	return Base{
		Path:     path,
		Title:    translate.Text(title),
		Summary:  translate.Text(summary),
		Icon:     icon,
		Keywords: translate.Texts(keywords),
		Links:    map[Relation][]Link{Self: {Link{Href: path}}},
	}
}

func adjustIcon(icon string) string {
	icon = strings.TrimSpace(icon)
	if icon == "" {
		return ""
	} else if strings.HasPrefix(icon, "http://") || strings.HasPrefix(icon, "https://") || strings.HasPrefix(icon, "data:") || strings.HasPrefix(icon, "/") {
		return icon
	} else {
		return "/img/" + icon
	}
}

func (this Base) OmitFromSearch() bool {
	return false
}

func (this Base) GetBase() Base {
	return this
}

func (this Base) AddAction(title string, icon string, relation Relation, actionId string) {
	var url = this.Path
	if "" != actionId {
		url = url + "?action=" + actionId // TODO: Escape?
	}
	this.Links[relation] = append(this.Links[relation], Link{Href: url, Title: title, Icon: icon})
}

func (this Base) FirstAction() *Link {
	if len(this.Links[OrgRefudeAction]) > 0 {
		return &this.Links[OrgRefudeAction][0]
	} else {
		return nil
	}
}

func (this Base) SecondaryActions() []Link {
	if len(this.Links[OrgRefudeAction]) > 1 {
		return this.Links[OrgRefudeAction][1:]
	} else {
		return nil
	}
}

func (this Base) DeleteAction() *Link {
	if len(this.Links[OrgRefudeDelete]) > 0 {
		return &this.Links[OrgRefudeDelete][0]
	} else {
		return nil
	}
}

func (this Base) DoGet(w http.ResponseWriter, r *http.Request, o any) {
	respond.AsJson(w, o)
}

type Link struct {
	Href  string `json:"href"`
	Title string `json:"title,omitempty"`
	Icon  string `json:"icon,omitempty"`
	Type  string `json:"type,omitempty"`
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

type Getable interface {
	DoGet(w http.ResponseWriter, r *http.Request, o any)
}

type Postable interface {
	DoPost(string) (bool, error)
}

type Deleteable interface {
	DoDelete() error
}
