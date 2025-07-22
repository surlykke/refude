// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package entity

import (
	"github.com/surlykke/refude/internal/lib/icon"
	"github.com/surlykke/refude/internal/lib/link"
	"github.com/surlykke/refude/internal/lib/mediatype"
	"github.com/surlykke/refude/internal/lib/relation"
	"github.com/surlykke/refude/internal/lib/translate"
)

type Servable interface {
	GetBase() *Base
	OmitFromSearch() bool
}

type Base struct {
	Path      string              `json:"-"`
	Title     string              `json:"title"`
	Subtitle  string              `json:"subtitle,omitempty"`
	Icon      icon.Name           `json:"icon"`
	MediaType mediatype.MediaType `json:"mediatype"`
	Links     []link.Link         `json:"links"`
	Keywords  []string            `json:"keywords"`
	Actions   []Action            `json:"actions"`
}

type Action struct {
	Id   string
	Name string
	Icon icon.Name
}

func (a Action) Href(path string) string {
	if a.Id != "" {
		path = path + "?action=" + a.Id
	}
	return path
}

func MakeBase(title string, subtitle string, icon icon.Name, mediatype mediatype.MediaType, keywords ...string) *Base {
	return &Base{
		Title:     translate.Text(title),
		Subtitle:  translate.Text(subtitle),
		Icon:      icon,
		MediaType: mediatype,
		Keywords:  translate.Texts(keywords),
	}
}

func (this *Base) OmitFromSearch() bool {
	return false
}

func (this *Base) GetBase() *Base {
	return this
}

// ------------ Don't call after published ------------------

func (this *Base) BuildLinks() {
	this.Links = make([]link.Link, 1+len(this.Actions), 1+len(this.Actions))
	this.Links[0] = link.Link{Href: this.Path, Title: this.Title, Icon: this.Icon, Relation: relation.Self}
	for i, action := range this.Actions {
		this.Links[i+1] = link.Link{Href: action.Href(this.Path), Title: action.Name, Relation: relation.Action}
	}
}

func (this *Base) AddAction(id string, name string, iconUrl icon.Name) {
	this.Actions = append(this.Actions, Action{Id: id, Name: translate.Text(name)})
}

func (this *Base) ActionLinks() []link.Link {
	var actionLinks = make([]link.Link, 0, len(this.Links))
	for _, l := range this.Links {
		if l.Relation == relation.Action || l.Relation == relation.Delete {
			actionLinks = append(actionLinks, l)
		}
	}
	return actionLinks
}

/*func (this *ResourceData) AddDeleteAction(actionId string, title string, comment string, iconName icon.Name) {
	this.Links = append(this.Links, Link{Href: href.Of(this.Path).P("action", actionId), Title: title, Comment: comment, Icon: iconName, Relation: relation.Delete})
}*/
