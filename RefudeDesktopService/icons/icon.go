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

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/slice"
)

type IconResource struct{}

func (ir IconResource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if names := r.URL.Query()["name"]; len(names) == 0 {
			respond.UnprocessableEntity(w, errors.New("no name given"))
		} else {
			names = dashSplit(names)
			var themeId = requests.GetSingleQueryParameter(r, "theme", "hicolor")
			var size = uint64(32)
			var err error
			if len(r.URL.Query()["size"]) > 0 {
				size, err = strconv.ParseUint(r.URL.Query()["size"][0], 10, 32)
				if err != nil {
					respond.UnprocessableEntity(w, errors.New("Invalid size given:"+r.URL.Query()["size"][0]))
				}
			}

			if image, ok := findImage(themeId, uint32(size), names...); !ok {
				w.WriteHeader(http.StatusNotFound)
			} else {
				http.ServeFile(w, r, image.Path)
			}
		}
	} else {
		respond.NotAllowed(w)
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

func findImage(themeId string, size uint32, iconNames ...string) (IconImage, bool) {
	lock.Lock()
	defer lock.Unlock()

	var idsToVisit = []string{themeId}
	for i := 0; i < len(idsToVisit); i++ {
		if theme, ok := themes["/icontheme/"+idsToVisit[i]]; ok {
			if imageList, ok := findImageListInMap(themeIcons[idsToVisit[i]], iconNames...); ok {
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

	if imageList, ok := findImageListInMap(themeIcons["hicolor"], iconNames...); ok {
		return findBestMatch(imageList, size), true
	}

	if imageList, ok := findImageListInMap(sessionIcons, iconNames...); ok {
		return findBestMatch(imageList, size), true
	}

	image, ok := findImageInMap(otherIcons, iconNames...)
	return image, ok
}

func findImageListInMap(m map[string][]IconImage, iconNames ...string) ([]IconImage, bool) {
	for _, iconName := range iconNames {
		if list, ok := m[iconName]; ok {
			return list, true
		}
	}
	return nil, false
}

func findImageInMap(m map[string]IconImage, iconNames ...string) (IconImage, bool) {
	for _, iconName := range iconNames {
		if image, ok := m[iconName]; ok {
			return image, true
		}
	}

	return IconImage{}, false
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
