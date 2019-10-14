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
	Self         string                    `json:"_self,omitempty"`
	RefudeType   string                    `json:"_refudetype,omitempty"`
	Actions      map[string]ResourceAction `json:"_actions"`
	DeleteAction *DeleteAction             `json:"_delete,omitempty"`
}

func (l *Links) Init(self string, refudeType string) {
	l.Self = self
	l.RefudeType = refudeType
	l.Actions = make(map[string]ResourceAction)
}

type Executer func()

type ResourceAction struct {
	Description string
	IconName    string
	Executer    Executer `json:"-"`
}

func (l *Links) GetSelf() string {
	return l.Self
}

type DeleteAction struct {
	Description string
	Executer    Executer `json:"-"`
}

func (l *Links) AddAction(actionId string, action ResourceAction) {
	l.Actions[actionId] = action
}

func (l *Links) SetDeleteAction(deleteAction *DeleteAction) {
	l.DeleteAction = deleteAction
}

func (l *Links) POST(w http.ResponseWriter, r *http.Request) {
	if l.Actions == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		var actionId = requests.GetSingleQueryParameter(r, "action", "default")
		if action, ok := l.Actions[actionId]; ok {
			action.Executer()
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
	}
}

func (l *Links) DELETE(w http.ResponseWriter, r *http.Request) {
	if l.DeleteAction == nil {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		l.DeleteAction.Executer()
		w.WriteHeader(http.StatusAccepted)
	}
}

func (l *Links) PATCH(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}
