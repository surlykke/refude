// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"math"

	"github.com/surlykke/RefudeServices/lib/slice"
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
