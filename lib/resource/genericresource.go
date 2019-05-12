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

type Executer func()

type ResourceAction struct {
	Description string
	IconName    string
	Executer    Executer `json:"-"`
}

// For embedding
type GeneralTraits struct {
	Self       string                    `json:"_self,omitempty"`
	RefudeType string                    `json:"_refudetype,omitempty"`
	Actions    map[string]ResourceAction `json:"_actions,omitempty"`
}

func (gt *GeneralTraits) AddAction(actionId string, action ResourceAction) {
	if gt.Actions == nil {
		gt.Actions = make(map[string]ResourceAction)
	}
	gt.Actions[actionId] = action
}

func (gt *GeneralTraits) POST(w http.ResponseWriter, r *http.Request) {
	if gt.Actions == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		var actionId = requests.GetSingleQueryParameter(r, "action", "default")
		if action, ok := gt.Actions[actionId]; ok {
			action.Executer()
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
	}
}

func (gt *GeneralTraits) GetSelf() string {
	return gt.Self
}
