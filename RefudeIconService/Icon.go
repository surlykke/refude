package main

import "math"

type Image struct {
	Context string
	MinSize uint32
	MaxSize uint32
	Path    string
}

type Icon struct {
	Name    string
	Images  []Image
}

func (icon *Icon) AddImage(image Image, check bool) {
	if check {
		for _, img := range icon.Images {
			if img.Path == image.Path  {
				return
			}
		}
	}

	icon.Images = append(icon.Images, image)
}


// Somewhat inspired by pseudocode example in
//    https://specifications.freedesktop.org/icon-theme-spec/icon-theme-spec-latest.html
// Caller should guarantee that icon's list of images is not empty
// Returns Image
func (icon *Icon) FindImage(size uint32) Image {
	shortestDistanceSoFar := uint32(math.MaxUint32)
	candidateIndex := -1

	for index, image := range icon.Images {
		var distance uint32
		if image.MinSize > size {
			distance = image.MinSize - size
		} else if image.MaxSize < size {
			distance = size - image.MaxSize
		} else {
			distance = 0
		}

		if distance < shortestDistanceSoFar {
			shortestDistanceSoFar = distance
			candidateIndex = index
		}
		if distance == 0 {
			break
		}
	}

	if candidateIndex < 0 {
		return Image{}
	} else {
		return icon.Images[candidateIndex]
	}

}
