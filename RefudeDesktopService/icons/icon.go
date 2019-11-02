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

	"github.com/surlykke/RefudeServices/lib/slice"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
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
	Path    string
}

type IconResource struct {
	resource.Links
}

func (ir *IconResource) GET(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else if name := requests.GetSingleQueryParameter(r, "name", ""); name == "" {
		requests.ReportUnprocessableEntity(w, errors.New("no name given"))
	} else {
		var themeId = requests.GetSingleQueryParameter(r, "theme", "hicolor")
		var size = uint64(32)
		var err error
		if len(r.URL.Query()["size"]) > 0 {
			size, err = strconv.ParseUint(r.URL.Query()["size"][0], 10, 32)
			if err != nil {
				requests.ReportUnprocessableEntity(w, errors.New("Invalid size given:"+r.URL.Query()["size"][0]))
			}
		}

		if image, ok := findImage(themeId, name, uint32(size)); !ok {
			w.WriteHeader(http.StatusNotFound)
		} else {
			http.ServeFile(w, r, image.Path)
		}
	}
}

func findImage(themeId string, iconName string, size uint32) (IconImage, bool) {
	lock.Lock()
	defer lock.Unlock()

	var idsToVisit = []string{themeId}
	for i := 0; i < len(idsToVisit); i++ {
		if theme, ok := themes[idsToVisit[i]]; ok {
			if imageList, ok := themeIcons[idsToVisit[i]][iconName]; ok {
				return findBestMatch(imageList, size), true
			} else {
				for _, parentThemeId := range theme.Inherits {
					if parentThemeId != "hicolor" {
						idsToVisit = slice.AppendIfNotThere(idsToVisit, parentThemeId)
					}
				}
			}
		}
	}

	if imageList, ok := themeIcons["hicolor"][iconName]; ok {
		return findBestMatch(imageList, size), true
	}

	if imageList, ok := sessionIcons[iconName]; ok {
		return findBestMatch(imageList, size), true
	}

	image, ok := otherIcons[iconName]
	return image, ok
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
