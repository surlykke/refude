// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"github.com/surlykke/RefudeServices/lib/resource"
)

const WindowMediaType resource.MediaType = "application/vnd.org.refude.wmwindow+json"

type Window struct {
	resource.AbstractResource
	Id            uint32
	Parent        uint32
	StackOrder    int
	X,Y,W,H       int
	Name          string
	IconName      string `json:",omitempty"`
	States        []string
}

