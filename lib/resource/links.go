// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/requests"
)

// For embedding
type Links struct {
	Self       string `json:"_self,omitempty"`
	RefudeType string `json:"_refudetype,omitempty"`
}

func MakeLinks(self string, refudeType string) Links {
	return Links{
		Self:       self,
		RefudeType: refudeType,
	}
}

type Actions struct {
	PostActions  map[string]ResourceAction `json:"_post,omitempty"`
	DeleteAction *ResourceAction           `json:"_delete,omitempty"`
	PatchAction  *ResourceAction           `json:"_patch,omitempty"`
}

type ResourceAction struct {
	Description string
	IconName    string
	Executer    Executer `json:"-"`
}

type Executer func()

func (a *Actions) SetPostAction(actionId string, action ResourceAction) {
	if a.PostActions == nil {
		a.PostActions = make(map[string]ResourceAction)
	}
	a.PostActions[actionId] = action
}

func (a *Actions) POST(w http.ResponseWriter, r *http.Request) {
	if a.PostActions == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		var actionId = requests.GetSingleQueryParameter(r, "action", "default")
		if action, ok := a.PostActions[actionId]; ok {
			action.Executer()
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
	}
}

func (a *Actions) DELETE(w http.ResponseWriter, r *http.Request) {
	if a.DeleteAction == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		a.DeleteAction.Executer()
		w.WriteHeader(http.StatusAccepted)
	}
}

func (a *Actions) PATCH(w http.ResponseWriter, r *http.Request) {
	if a.PatchAction == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		a.PatchAction.Executer()
		w.WriteHeader(http.StatusAccepted)
	}
}
