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

type Action2 struct {
	Description   string
	IconName  	  string
	executer      Executer
}

func MakeAction2(Description string, IconName string, executer Executer) *Action2 {
	return &Action2{Description, IconName, executer}
}

func (a *Action2) POST(w http.ResponseWriter, r *http.Request) {
	if a.executer != nil {
		a.executer()
	}
	w.WriteHeader(http.StatusAccepted)
}

