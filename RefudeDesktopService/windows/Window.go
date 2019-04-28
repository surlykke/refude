// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"log"
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/lib/resource"
)

const WindowMediaType resource.MediaType = "application/vnd.org.refude.wmwindow+json"

type Window struct {
	resource.GenericResource
	Id         uint32
	Parent     uint32
	StackOrder int
	X, Y       int32
	W, H       uint32
	Name       string
	IconName   string `json:",omitempty"`
	States     []string
}

func windowSelf(windowId uint32) resource.StandardizedPath {
	return resource.Standardizef("/window/%d", windowId)
}

type WindowCollection struct{}

func (wc WindowCollection) Get(path string) resource.Resource {
	if !strings.HasPrefix(path, "/window/") {
		return nil
	} else if id, err := strconv.ParseUint(string(path[len("/window/"):]), 10, 32); err != nil {
		return nil
	} else {
		window, err := getWindow(uint32(id))
		if err != nil {
			return nil
		}
		return window
	}
}

func (wc WindowCollection) GetList(path string) []resource.Resource {
	if "/windows" != path {
		return nil
	} else if windows, err := getWindows(); err != nil {
		log.Printf("Error getting windows: %v\n", err)
		return nil
	} else {
		var result = make([]resource.Resource, len(windows), len(windows))
		for i := 0; i < len(windows); i++ {
			result[i] = windows[i]
		}
		return result
	}

}

var Windows = WindowCollection{}
