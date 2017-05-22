/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package main

import (
	"net/http"
	"github.com/surlykke/RefudeServices/common"
)



type PowerAction struct {
	Id string
	Name string
	Comment string
	IconName string
	Can bool
}

func NewPowerAction(Id string, Name string, Comment string, IconName string) PowerAction {
	bool can 	
}

func (p PowerAction) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if (r.Method == "GET") {
		common.ServeAsJson(w, r, p)
	} else if r.Method == "POST" {
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}


