// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package action

import (
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"net/http"
	"github.com/surlykke/RefudeServices/lib/resource"
)

const ActionMediaType mediatype.MediaType = "application/vnd.org.refude.action+json"

type Executer func()

type Action struct {
	resource.AbstractResource
	Name     string
	Comment  string
	IconName string
	Hint     string
	RelevanceHint int64
	executer Executer
}

func MakeAction(Self string, Name string, Comment string, IconName string, hint string, executer Executer) *Action {
	var act = Action{}
	act.Self = Self
	act.Mt = ActionMediaType
	act.Name= Name
	act.Comment = Comment
	act.IconName = IconName
	act.Hint = hint
	act.executer=executer
	return &act
}

func (a *Action) POST(w http.ResponseWriter, r *http.Request) {
	if a.executer != nil {
		a.executer()
	}
	w.WriteHeader(http.StatusAccepted)
}
