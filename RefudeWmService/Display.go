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
	"github.com/surlykke/RefudeServices/lib/stringlist"
)

type Rect struct {
	X,Y int16
	W,H uint16
}

type Display struct {
	W,H uint16
	Screens []Rect
}

func (d* Display) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stringlist.ServeGetAsJson(w, r, d)
}
