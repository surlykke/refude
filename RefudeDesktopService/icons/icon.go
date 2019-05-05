// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"github.com/surlykke/RefudeServices/lib/resource"
)

type Icon struct {
	resource.GenericResource
	Name   string
	Theme  string
	Images []IconImage
}

type IconImage struct {
	Type    string
	Context string
	MinSize uint32
	MaxSize uint32
	Path    string
}

type Theme struct {
	resource.GenericResource
	Id       string
	Name     string
	Comment  string
	Inherits []string
	Dirs     map[string]IconDir
}

type IconDir struct {
	Path    string
	MinSize uint32
	MaxSize uint32
	Context string
}

type PngSvgPair struct {
	Png *Icon
	Svg *Icon
}
