// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
)

type Icon struct {
	Name   string
	Theme  string
	Images []IconImage
}

type IconImage struct {
	Context string
	MinSize uint32
	MaxSize uint32
	Data    []byte // Only one of
	Path    string // these two non-zero
}

type IconMap map[string]*Icon // Maps icon name -> Icon

func (i *Icon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var size = uint64(32)
		var err error
		if len(r.URL.Query()["size"]) > 0 {
			size, err = strconv.ParseUint(r.URL.Query()["size"][0], 10, 32)
			if err != nil {
				respond.UnprocessableEntity(w, errors.New("Invalid size given:"+r.URL.Query()["size"][0]))
			}
		}

		var image = findBestMatch(i.Images, uint32(size))
		image.ServeHTTP(w, r)
	} else {
		respond.NotAllowed(w)
	}
}

func findBestMatch(images []IconImage, size uint32) IconImage {
	var shortestDistanceSoFar = uint32(math.MaxUint32)
	var candidate IconImage

	for _, img := range images {
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

	return candidate
}

func (ii IconImage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ii.Data != nil {
		respond.AsPng(w, ii.Data)
	} else {
		http.ServeFile(w, r, ii.Path)
	}
}

/**
 * By the icon naming specification, dash ('-') seperates 'levels of specificity'. So given an icon name
 * 'input-mouse-usb', the levels of spcicificy, and the names and order we search will be: 'input-mouse-usb',
 * 'input-mouse' and 'input'
 */
func dashSplit(names []string) []string {
	var res = make([]string, 0, len(names)*2)
	for _, name := range names {
		for {
			res = append(res, name)
			if pos := strings.LastIndex(name, "-"); pos > 0 {
				name = name[0:pos]
			} else {
				break
			}
		}
	}
	return res
}
