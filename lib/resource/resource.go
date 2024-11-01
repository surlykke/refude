// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package resource

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/path"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/respond"
)

type Resource interface {
	Data() *ResourceData
	OmitFromSearch() bool
}

type ResourceData struct {
	Path          path.Path `json:"self"`
	Title         string    `json:"title"`
	Comment       string    `json:"comment,omitempty"`
	Icon          icon.Name
	DefaultAction string
	DeleteAction  string
	DeleteIcon    icon.Name
	Actions       []Action
	Type          mediatype.MediaType `json:"type"`
	Keywords      []string            `json:"keywords"`
}

func MakeBase(path path.Path, title, comment string, icon icon.Name, mType mediatype.MediaType) *ResourceData {
	var br = ResourceData{
		Path:    path,
		Title:   title,
		Comment: comment,
		Icon:    icon,
		Type:    mType,
	}
	return &br
}

func (this *ResourceData) Data() *ResourceData {
	return this
}

func (this *ResourceData) Link() Link {
	return LinkTo(this)
}

func (this *ResourceData) OmitFromSearch() bool {
	return false
}

func (this *ResourceData) GetActionLinks() []Link {
	var result = make([]Link, 0, 8)
	if this.DefaultAction != "" {
		result = append(result, Link{Path: this.Path, Title: this.DefaultAction, Icon: this.Icon, Relation: relation.DefaultAction})
	}
	for _, a := range this.Actions {
		result = append(result, Link{Path: path.Of(this.Path, "?action=", a.Id), Title: a.Title, Icon: a.Icon, Relation: relation.Action})
	}
	if this.DeleteAction != "" {
		result = append(result, Link{Path: this.Path, Title: this.DeleteAction, Icon: this.DeleteIcon, Relation: relation.Delete})
	}

	return result
}

var httpLocalHost7838 = []byte("http://localhost:7938")

// --------------------- Link --------------------------------------

type Link struct {
	Path     path.Path           `json:"href"`
	Title    string              `json:"title,omitempty"`
	Comment  string              `json:"comment,omitempty"`
	Icon     icon.Name           `json:"icon,omitempty"`
	Relation relation.Relation   `json:"rel,omitempty"`
	Type     mediatype.MediaType `json:"type,omitempty"`
	// --- Used for search -------
	Keywords []string `json:"-"`
	Rank     int      `json:"-"`
}

func LinkTo(res Resource) Link {
	return Link{Path: res.Data().Path, Title: res.Data().Title, Comment: res.Data().Comment, Icon: res.Data().Icon, Relation: relation.Related, Type: res.Data().Type}
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
