// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package entity

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/surlykke/refude/internal/lib/response"
	"github.com/surlykke/refude/internal/lib/translate"
)

type Servable interface {
	GetBase() *Base
	OmitFromSearch() bool
}

type Base struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
	Icon     string `json:"icon"`
	Kind     string `json:"type"`
	Meta     Meta   `json:"links"`
}

type Meta struct {
	Path     string
	Actions  []Action
	Keywords []string // TODO Maybe a function, including keywords from actions
}

func (this *Meta) MarshalJSON() ([]byte, error) {
	return json.Marshal(buildLinks(this))
}

type Action struct {
	Id   string
	Name string
	Icon string
}

func MakeBase(title string, subtitle string, icon string, kind string, keywords ...string) *Base {
	icon = adjustIcon(icon)
	return &Base{
		Title:    translate.Text(title),
		Subtitle: translate.Text(subtitle),
		Icon:     icon,
		Kind:     kind,
		Meta:     Meta{Keywords: translate.Texts(keywords)},
	}
}

func adjustIcon(icon string) string {
	if strings.HasPrefix(icon, "http://") || strings.HasPrefix(icon, "https://") || strings.HasPrefix(icon, "/icon?name=") {
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

func (this *Base) Links(rel ...Relation) []Link {
	return buildLinks(&this.Meta)
}

func buildLinks(meta *Meta, rel ...Relation) []Link {
	var links = make([]Link, 0, 1+len(meta.Actions))
	if len(rel) == 0 || slices.Index(rel, Self) > -1 {
		links = append(links, Link{Href: meta.Path, Relation: Self})
	}
	if len(rel) == 0 || slices.Index(rel, OrgRefudeAction) > -1 {
		for _, action := range meta.Actions {
			links = append(links, Link{Href: meta.Path, Title: action.Name, Icon: action.Icon, Relation: OrgRefudeAction})
		}
	}
	return links
}

func (this *Base) AddAction(id string, name string, icon string) {
	icon = adjustIcon(icon)
	this.Meta.Actions = append(this.Meta.Actions, Action{Id: id, Name: translate.Text(name)})
}

/*func (this *ResourceData) AddDeleteAction(actionId string, title string, comment string, iconName icon.Name) {
	this.Links = append(this.Links, Link{Href: href.Of(this.Path).P("action", actionId), Title: title, Comment: comment, Icon: iconName, Relation: entity.Delete})
}*/

type Link struct {
	Href     string   `json:"href"`
	Title    string   `json:"title,omitempty"`
	Icon     string   `json:"icon,omitempty"`
	Relation Relation `json:"rel,omitempty"`
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
	DoPost(string) response.Response
}

type Deleteable interface {
	DoDelete() response.Response
}
