// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/lib/resource"
)

type Window struct {
	resource.Links
	resource.Actions
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

func (wc WindowCollection) Get(path string) interface{} {
	if path == "/windows" {
		if windows, err := getWindows(); err != nil {
			return nil
		} else {
			return windows
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
		return window
	}
}

var Windows = resource.MakeJsonResourceServer(WindowCollection{})
