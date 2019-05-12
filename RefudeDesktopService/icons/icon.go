// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"fmt"
	"math"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
)

type Icon struct {
	resource.GeneralTraits
	Name   string
	Theme  string
	Images ImageList
}

type IconImage struct {
	Type    string
	Context string
	MinSize uint32
	MaxSize uint32
	Path    string
}

type ImageList []IconImage

type Theme struct {
	resource.GeneralTraits
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

type IconImgResource struct {
	images ImageList
}

func (iir IconImgResource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var size = uint32(32)

	if len(r.URL.Query()["size"]) > 0 {
		var ok bool
		if size, ok = readUint32(r.URL.Query()["size"][0]); !ok {
			requests.ReportUnprocessableEntity(w, fmt.Errorf("Invalid size"))
			return
		}
	}

	var shortestDistanceSoFar = uint32(math.MaxUint32)
	var candidate IconImage

	for _, img := range iir.images {
		var distance uint32
		if img.MinSize > size {
			distance = img.MinSize - size
		} else if img.MaxSize < size {
			distance = size - img.MaxSize
		} else {
			distance = 0
		}

		if distance < shortestDistanceSoFar {
			shortestDistanceSoFar = distance
			candidate = img
		}
		if distance == 0 {
			break
		}

	}

	http.ServeFile(w, r, candidate.Path)
}
