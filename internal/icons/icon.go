package icons

import (
	"math"

	"github.com/surlykke/refude/internal/lib/resource"
)

const iconPrefix = "/icon/"

type Img struct {
	OsPath  string // Path
	MaxSize uint32
	MinSize uint32
}

type Icon struct {
	resource.Base
	Name  string
	Theme string
	Imgs  []Img
}

func (this *Icon) HasPath(osPath string) bool {
	for _, img := range this.Imgs {
		if img.OsPath == osPath {
			return true
		}
	}
	return false
}

func (this *Icon) GetPath(size uint32) string {
	var shortestDistanceSoFar = uint32(math.MaxUint32)
	var path = ""
	for _, img := range this.Imgs {
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
			path = img.OsPath
			shortestDistanceSoFar = distance
		}
		if shortestDistanceSoFar == 0 {
			break
		}
	}
	return path
}
