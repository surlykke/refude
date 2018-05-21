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
	"net/http"
	"strconv"
	"github.com/surlykke/RefudeServices/lib/requestutils"
)

const WindowMediaType mediatype.MediaType = "application/vnd.org.refude.wmwindow+json"

type Window struct {
	resource.AbstractResource
	Id            xproto.Window
	X, Y, H, W    int
	Name          string
	IconName      string        `json:",omitempty"`
	States        []string
	RelevanceHint int64
}

func (win *Window) POST(w http.ResponseWriter, r *http.Request) {
	var value = requestutils.GetSingleQueryParameter(r, "opacity", "1")
	if opacity, err := strconv.ParseFloat(value, 64); err != nil || opacity < 0 || opacity > 1 {
		requestutils.ReportUnprocessableEntity(w, []byte("parameter 'opacity' must be a float between 0 and 1"))
	} else {
		highlightRequests <- highlightRequest{win.Id, opacity}
	}
}
