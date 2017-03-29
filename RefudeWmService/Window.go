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
	"fmt"
)

var xUtil       *xgbutil.XUtil

type Window struct {
	Id      xproto.Window
	X,Y,H,W int
	Name    string
	IconUrl string
	States  common.StringList
	Actions map[string]*Action
}

type Action struct {
	Name    string
	Comment string
	IconUrl string
	winId   xproto.Window
	X,Y,W,H int
	States  []string
}

func (win Window) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		common.ServeAsJson(w, r, win)
	} else if r.Method == "POST" {
		if actionv,ok := r.URL.Query()["action"]; ok && len(actionv) > 0 && actionv[0] != "_default" {
			w.WriteHeader(http.StatusNotAcceptable)
		} else {
			ewmh.ActiveWindowReq(xUtil, xproto.Window(win.Id))
			w.WriteHeader(http.StatusAccepted)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (ac* Action) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		common.ServeAsJson(w, r, ac)
	} else if r.Method == "POST" {
		fmt.Println("POST mod ac.winId: ", ac.winId)
		if err := ewmh.ActiveWindowReq(xUtil, ac.winId); err == nil {
			w.WriteHeader(http.StatusAccepted)
		} else {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
