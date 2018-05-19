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

type Window struct {
	resource.AbstractResource
	Id            xproto.Window
	X, Y, H, W    int
	Name          string
	IconName      string        `json:",omitempty"`
	States        []string
	RelevanceHint int
}

