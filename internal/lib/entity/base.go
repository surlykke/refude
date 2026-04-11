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
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
	Icon     string `json:"icon"`
	Kind     string `json:"type"`
	Path     string `json:"path"`
	Links    map[Relation][]Link
	Keywords []string
	Actions  []Action `json:"-"`
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
		Keywords: translate.Texts(keywords),
		Links:    make(map[Relation][]Link),
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

// Call before serving
func (this *Base) SetPath(path string) {
	this.Path = path
	this.Links = make(map[Relation][]Link)
	this.Links[Self] = []Link{{Href: path}}
	for _, a := range this.Actions {
		this.Links[OrgRefudeAction] = append(this.Links[OrgRefudeAction], Link{Href: path + "?action=" + a.Id, Title: a.Name, Icon: a.Icon})
	}
}

func (this *Base) GetLinks(rel ...Relation) []Link {
	var tmp = []Link{}
	for _, r := range rel {
		if l, ok := this.Links[r]; ok {
			tmp = append(tmp, l...)
		}
	}

	return tmp
}

func (this *Base) AddAction(id string, name string, icon string) {
	icon = adjustIcon(icon)
	this.Actions = append(this.Actions, Action{Id: id, Name: translate.Text(name)})
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
	DoPost(string) (bool, error)
}

type Deleteable interface {
	DoDelete() error
}
