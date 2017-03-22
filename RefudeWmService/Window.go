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
	"github.com/surlykke/RefudeServices/common"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil"
)


type Window struct {
	x       *xgbutil.XUtil
	Id      WId
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
}

type WId xproto.Window

type WIdList []WId


func (w *Window) Data(r *http.Request) (int, string, []byte) {
	if r.Method == "GET" {
		return common.GetJsonData(w)
	} else if r.Method == "POST" {
		if actionv,ok := r.URL.Query()["action"]; ok && len(actionv) > 0 && actionv[0] != "_default" {
			return http.StatusNotAcceptable, "", nil
		} else {
			ewmh.ActiveWindowReq(w.x, xproto.Window(w.Id))
			return http.StatusAccepted, "", nil
		}
	} else {
		return http.StatusMethodNotAllowed, "", nil
	}
}

func (w *Window) Equal( w2 *Window) bool {
	if w == w2 {
		return true
	} else if w == nil || w2 == nil {
		return false
	} else if w.Id != w2.Id || w.X != w2.X || w.Y != w2.Y || w.H != w2.H || w.W != w2.W || w.IconUrl != w2.IconUrl {
		return false
	} else if len(w.States) != len(w2.States) {
		return false
	} else {
		for i, state := range w.States {
			if state != w2.States[i] {
				return false
			}
		}
	}

	return true
}


func (wl WIdList) Equal(wl2 WIdList) bool {
	if len(wl) != len(wl2) {
		return false
	} else {
		for i := 0; i < len(wl); i++ {
			if wl[i] != wl2[i] {
				return false
			}
		}
	}
	return true
}

