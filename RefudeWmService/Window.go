// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/resource"
)

const WindowMediaType mediatype.MediaType = "application/vnd.org.refude.wmwindow+json"
const DisplayMediaType mediatype.MediaType = "application/vnd.org.refude.wmdisplay+json"

type Rect struct {
	X, Y int
	W, H uint
}

type Window struct {
	resource.AbstractResource
	Id            xproto.Window
	Geometry      Rect
	Name          string
	IconName      string `json:",omitempty"`
	States        []string
	RelevanceHint int64
}

type Display struct {
	resource.AbstractResource
	RootGeometry Rect
	Screens      Screens
}

type Screen struct {
	X, Y int
	W, H uint
}

type Screens []Screen

func (s Screens) Len() int {
	return len(s)
}

func (s Screens) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Screens) Less(i, j int) bool {
	return s[i].X < s[j].X ||
		s[i].X == s[j].X && s[i].Y < s[j].Y ||
		s[i].X == s[j].X && s[i].Y == s[j].Y && s[i].W < s[j].W ||
		s[i].X == s[j].X && s[i].Y == s[j].Y && s[i].W == s[j].W && s[i].H < s[j].H
}
