// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"net/http"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/BurntSushi/xgbutil/ewmh"
)

const WindowMediaType mediatype.MediaType = "application/vnd.org.refude.wmwindow+json"

type Window struct {
	x             *xgbutil.XUtil
	Id            xproto.Window
	X, Y, H, W    int
	Name          string
	IconName      string        `json:",omitempty"`
	IconUrl       string        `json:",omitempty"`
	States        []string
	Actions       map[string]Action
	RelevanceHint int
	Self          string
}

type Action struct {
	Name    string
	Comment string
}

func (win *Window) POST(w http.ResponseWriter, r *http.Request) {
	if actionv, ok := r.URL.Query()["action"]; ok && len(actionv) > 0 && actionv[0] != "_default" {
		w.WriteHeader(http.StatusNotAcceptable)
	} else {
		ewmh.ActiveWindowReq(win.x, xproto.Window(win.Id))
		w.WriteHeader(http.StatusAccepted)
	}
}

