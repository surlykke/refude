// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/lib/resource"
)

const WindowMediaType resource.MediaType = "application/vnd.org.refude.wmwindow+json"

type Window struct {
	resource.GeneralTraits
	resource.DefaultMethods
	Id         uint32
	Parent     uint32
	StackOrder int
	X, Y       int32
	W, H       uint32
	Name       string
	IconName   string `json:",omitempty"`
	States     []string
}

func windowSelf(windowId uint32) string {
	return fmt.Sprintf("/window/%d", windowId)
}

type WindowCollection struct{}

type WindowList []*Window

func (wl WindowList) ServeHttp(w http.ResponseWriter, r *http.Request) {
	var bytes, etag = resource.ToBytesAndEtag(wl)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", etag)
	_, _ = w.Write(bytes)
}

func (wc WindowCollection) Get(path string) resource.Resource {
	if path == "/windows" {
		if windows, err := getWindows(); err != nil {
			return nil
		} else {
			return WindowList(windows)
		}
	} else if !strings.HasPrefix(path, "/window/") {
		return nil
	} else if id, err := strconv.ParseUint(string(path[len("/window/"):]), 10, 32); err != nil {
		return nil
	} else {
		window, err := getWindow(uint32(id))
		if err != nil {
			return nil
		}
		return resource.MakeJsonResource(window)
	}
}

var Windows = WindowCollection{}
