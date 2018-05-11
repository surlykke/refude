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
	resource.Self
	Name             string
	Comment          string
	IconName         string
	PresentationHint string
	executer         Executer
}

func MakeAction(Name string, Comment string, IconName string, presentationHint string, executer Executer) *Action {
	return &Action{Name: Name, Comment: Comment, IconName: IconName, PresentationHint:presentationHint,  executer:executer}
}

func (a *Action) POST(w http.ResponseWriter, r *http.Request) {
	if a.executer != nil {
		a.executer()
	}
	w.WriteHeader(http.StatusAccepted)
}
