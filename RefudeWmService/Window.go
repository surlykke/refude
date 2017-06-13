/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package main

import (
	"github.com/BurntSushi/xgb/xproto"
	"net/http"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil"
	"github.com/surlykke/RefudeServices/lib/resource"
)


type Window struct {
	x       *xgbutil.XUtil
	Id      xproto.Window
	X,Y,H,W int
	Name    string
	IconUrl string
	States  []string
	Actions map[string]Action
}

type Action struct {
	Name    string
	Comment string
	IconUrl string
	X,Y,W,H int
}

func WindowPOST(this *resource.Resource, w http.ResponseWriter, r *http.Request) {
	win := this.Data.(Window)
	if actionv,ok := r.URL.Query()["action"]; ok && len(actionv) > 0 && actionv[0] != "_default" {
		w.WriteHeader(http.StatusNotAcceptable)
	} else {
		ewmh.ActiveWindowReq(win.x, xproto.Window(win.Id))
		w.WriteHeader(http.StatusAccepted)
	}
}

