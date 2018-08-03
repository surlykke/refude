// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package lib

import (
	"net/http"
)

const ActionMediaType MediaType = "application/vnd.org.refude.action+json"

type Executer func()

type Action struct {
	AbstractResource
	Name      string
	Comment   string
	IconName  string
	RelevanceHint int64
	executer  Executer
}


func MakeAction(Self StandardizedPath, Name string, Comment string, IconName string, executer Executer) *Action {
	var act = Action{}
	act.Self = Self
	act.Mt = ActionMediaType
	act.Name= Name
	act.Comment = Comment
	act.IconName = IconName
	act.executer=executer
	return &act
}

func (a *Action) POST(w http.ResponseWriter, r *http.Request) {
	if a.executer != nil {
		a.executer()
	}
	w.WriteHeader(http.StatusAccepted)
}

