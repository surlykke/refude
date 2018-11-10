// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"github.com/surlykke/RefudeServices/lib/requests"
	"log"
	"net/http"
)


type Action2 struct {
	Description   string
	IconName  	  string
	Executer      Executer `json:"-"`
}


func (a *Action2) POST(w http.ResponseWriter, r *http.Request) {
	if a.Executer != nil {
		a.Executer()
	}
}



type AbstractResource struct {
	Self StandardizedPath `json:"_self,omitempty"`
	Relates map[MediaType][]StandardizedPath `json:"_relates,omitempty"`
	Mt MediaType `json:"-"`
	Actions map[string]Action2 `json:"_actions,omitempty"`
}

func (ar *AbstractResource) GetSelf() StandardizedPath {
	return ar.Self
}

func (ar *AbstractResource) GetMt() MediaType {
	return ar.Mt
}

func Relate(r1, r2 *AbstractResource) {
	if r1.Self == "" || r2.Self == "" {
		log.Fatal("Relating resources with empty 'self'")
	}

	if r1.Relates == nil {
		r1.Relates = make(map[MediaType][]StandardizedPath)
	}
	if r2.Relates == nil {
		r2.Relates = make(map[MediaType][]StandardizedPath)
	}

	r1.Relates[r2.Mt] = append(r1.Relates[r2.Mt], r2.Self)
	r2.Relates[r1.Mt] = append(r2.Relates[r1.Mt], r1.Self)
}

func (ar *AbstractResource) AddAction(id string, description string, iconName string, executer Executer) {
	if ar.Actions == nil {
		ar.Actions = make(map[string]Action2)
	}
	ar.Actions[id] = Action2{description, iconName, executer}
}

func (ar *AbstractResource) POST(w http.ResponseWriter, r *http.Request) {
	var actionId = requests.GetSingleQueryParameter(r, "action", "default")
	if action, ok := ar.Actions[actionId]; ok {
		action.Executer()
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}


