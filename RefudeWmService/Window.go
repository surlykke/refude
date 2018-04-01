// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/BurntSushi/xgb/xproto"
	"net/http"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil"
	"github.com/surlykke/RefudeServices/lib/resource"
	"net/url"
	"strings"
)


type Window struct {
	x             *xgbutil.XUtil
	Id            xproto.Window
	X,Y,H,W       int
	Name          string
	IconUrl       string
	States        []string
	Actions       map[string]Action
	RelevanceHint int
	Self          string
}

type Action struct {
	Name    string
	Comment string
}

func (win *Window) GET(w http.ResponseWriter, r *http.Request) {
	resource.JsonGET(win, w)
}

func (win *Window) POST(w http.ResponseWriter, r *http.Request) {
	if actionv,ok := r.URL.Query()["action"]; ok && len(actionv) > 0 && actionv[0] != "_default" {
		w.WriteHeader(http.StatusNotAcceptable)
	} else {
		ewmh.ActiveWindowReq(win.x, xproto.Window(win.Id))
		w.WriteHeader(http.StatusAccepted)
	}
}

func Filter(resource interface{}, queryParams url.Values) bool {
	if w, ok := resource.(*Window); ok {
		if searchTerms, ok := queryParams["q"]; ok {
			for _,searchTerm := range searchTerms {
				if strings.Contains(strings.ToUpper(w.Name), strings.ToUpper(searchTerm)) {
					return true
				}
			}
		}
	}
	return false
}

